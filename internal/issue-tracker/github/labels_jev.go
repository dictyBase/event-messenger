package github

import (
	"context"
	"strings"
	"time"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	M "github.com/IBM/fp-go/v2/monoid"
	N "github.com/IBM/fp-go/v2/number"
	L "github.com/IBM/fp-go/v2/optics/lens"
	O "github.com/IBM/fp-go/v2/option"
	PR "github.com/IBM/fp-go/v2/pair"
	P "github.com/IBM/fp-go/v2/predicate"
	R "github.com/IBM/fp-go/v2/record"
	S "github.com/IBM/fp-go/v2/string"
	openrouter "github.com/OpenRouterTeam/go-sdk"
	"github.com/OpenRouterTeam/go-sdk/models/components"
	"github.com/OpenRouterTeam/go-sdk/models/operations"
	mapset "github.com/deckarep/golang-set/v2"
	arrutils "github.com/dictyBase/fp-go-loom/array"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	predarrays "github.com/dictyBase/fp-go-loom/predicate/array"
	predord "github.com/dictyBase/fp-go-loom/predicate/ord"
	"github.com/dictyBase/fp-go-loom/strutils"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/sirupsen/logrus"
)

const (
	defaultJevModel        = "~typesafe/jev-latest"
	jevConfidenceThreshold = 0.75
	jevRequestTimeout      = 10 * time.Second
	choiceNone             = "None"

	labelStrainOrder      = "Strain Order"
	labelPlasmidOrder     = "Plasmid Order"
	labelAxenicMedium     = "Axenic Medium"
	labelGoldenBraid      = "GoldenBraid"
	labelGWDI             = "GWDI Strain Order"
	labelBacterialFood    = "bacterial food source"
	labelMSI              = "Minority Serving Institution"
	labelResearchInst     = "Research Institution"
	labelUndergradCollege = "Undergrad  college"
	labelHighSchool       = "High School"
	labelCommercial       = "Commercial"
	labelCreditCard       = "Credit Card Pay"
	labelPurchaseOrder    = "PO Pay"
	labelWireTransfer     = "Wire Transfer Pay"
	labelChartString      = "Chart String Payment"
	labelWaiverNoPay      = "Waiver/ No payment"
	labelNoCharge         = "No Charge"

	questionInstitutionType = "institution_type"
	questionIsMSI           = "is_msi"
	questionPaymentLabel    = "payment_label"
	questionLocationLabel   = "location_label"
	questionBacterialFood   = "is_bacterial_food_source"
	questionGoldenBraid     = "is_goldenbraid"
	questionGWDI            = "is_gwdi"
	questionAxenicMedium    = "is_axenic_medium"
)

// DecisionsClient executes a TypeSafe System One Decisions request on OpenRouter.
type DecisionsClient interface {
	Create(
		ctx context.Context,
		request components.DecisionsRequest,
		opts ...operations.Option,
	) (*components.DecisionsResponse, error)
}

type labelResolveState struct {
	Client   DecisionsClient
	Logger   *logrus.Entry
	Model    string
	Data     *allData
	Order    *order.Order
	Base     []string
	Request  components.DecisionsRequest
	Response *components.DecisionsResponse
}

type noulLabelSpec struct {
	key   string
	label string
}

var (
	resolveClientLens = L.MakeLens(
		func(s labelResolveState) DecisionsClient { return s.Client },
		func(s labelResolveState, v DecisionsClient) labelResolveState {
			s.Client = v
			return s
		},
	)

	resolveLoggerLens = L.MakeLens(
		func(s labelResolveState) *logrus.Entry { return s.Logger },
		func(s labelResolveState, v *logrus.Entry) labelResolveState {
			s.Logger = v
			return s
		},
	)

	resolveModelLens = L.MakeLens(
		func(s labelResolveState) string { return s.Model },
		func(s labelResolveState, v string) labelResolveState {
			s.Model = v
			return s
		},
	)

	resolveDataLens = L.MakeLens(
		func(s labelResolveState) *allData { return s.Data },
		func(s labelResolveState, v *allData) labelResolveState {
			s.Data = v
			return s
		},
	)

	resolveOrderLens = L.MakeLens(
		func(s labelResolveState) *order.Order { return s.Order },
		func(s labelResolveState, v *order.Order) labelResolveState {
			s.Order = v
			return s
		},
	)

	resolveBaseLens = L.MakeLens(
		func(s labelResolveState) []string { return s.Base },
		func(s labelResolveState, v []string) labelResolveState {
			s.Base = v
			return s
		},
	)

	resolveReqLens = L.MakeLens(
		func(s labelResolveState) components.DecisionsRequest { return s.Request },
		func(s labelResolveState, v components.DecisionsRequest) labelResolveState {
			s.Request = v
			return s
		},
	)

	resolveRespLens = L.MakeLens(
		func(s labelResolveState) *components.DecisionsResponse { return s.Response },
		func(s labelResolveState, v *components.DecisionsResponse) labelResolveState {
			s.Response = v
			return s
		},
	)

	decReqModelLens = L.MakeLens(
		func(r components.DecisionsRequest) string { return r.Model },
		func(r components.DecisionsRequest, v string) components.DecisionsRequest {
			r.Model = v
			return r
		},
	)

	decReqStateLens = L.MakeLens(
		func(r components.DecisionsRequest) components.State { return r.State },
		func(r components.DecisionsRequest, v components.State) components.DecisionsRequest {
			r.State = v
			return r
		},
	)

	decReqQuestionsLens = L.MakeLens(
		func(r components.DecisionsRequest) map[string]components.Questions { return r.Questions },
		func(
			r components.DecisionsRequest,
			v map[string]components.Questions,
		) components.DecisionsRequest {
			r.Questions = v
			return r
		},
	)

	choiceQCriteriaLens = L.MakeLens(
		func(q components.DecisionsChoiceQuestion) map[string]*components.Criteria {
			return q.Criteria
		},
		func(
			q components.DecisionsChoiceQuestion,
			v map[string]*components.Criteria,
		) components.DecisionsChoiceQuestion {
			q.Criteria = v
			return q
		},
	)

	choiceQInstructionsLens = L.MakeLens(
		func(
			q components.DecisionsChoiceQuestion,
		) components.DecisionsChoiceQuestionInstructions {
			return q.Instructions
		},
		func(
			q components.DecisionsChoiceQuestion,
			v components.DecisionsChoiceQuestionInstructions,
		) components.DecisionsChoiceQuestion {
			q.Instructions = v
			return q
		},
	)

	choiceQTypeLens = L.MakeLens(
		func(q components.DecisionsChoiceQuestion) components.DecisionsChoiceQuestionType {
			return q.Type
		},
		func(
			q components.DecisionsChoiceQuestion,
			v components.DecisionsChoiceQuestionType,
		) components.DecisionsChoiceQuestion {
			q.Type = v
			return q
		},
	)

	noulQCriteriaLens = L.MakeLens(
		func(q components.DecisionsNoulQuestion) *components.DecisionsNoulQuestionCriteria {
			return q.Criteria
		},
		func(
			q components.DecisionsNoulQuestion,
			v *components.DecisionsNoulQuestionCriteria,
		) components.DecisionsNoulQuestion {
			q.Criteria = v
			return q
		},
	)

	noulQInstructionsLens = L.MakeLens(
		func(
			q components.DecisionsNoulQuestion,
		) components.DecisionsNoulQuestionInstructions {
			return q.Instructions
		},
		func(
			q components.DecisionsNoulQuestion,
			v components.DecisionsNoulQuestionInstructions,
		) components.DecisionsNoulQuestion {
			q.Instructions = v
			return q
		},
	)

	noulQTypeLens = L.MakeLens(
		func(q components.DecisionsNoulQuestion) components.DecisionsNoulQuestionType {
			return q.Type
		},
		func(
			q components.DecisionsNoulQuestion,
			v components.DecisionsNoulQuestionType,
		) components.DecisionsNoulQuestion {
			q.Type = v
			return q
		},
	)

	noulCriteriaTrueLens = L.MakeLens(
		func(c components.DecisionsNoulQuestionCriteria) components.True { return c.True },
		func(
			c components.DecisionsNoulQuestionCriteria,
			v components.True,
		) components.DecisionsNoulQuestionCriteria {
			c.True = v
			return c
		},
	)

	noulCriteriaFalseLens = L.MakeLens(
		func(c components.DecisionsNoulQuestionCriteria) components.False { return c.False },
		func(
			c components.DecisionsNoulQuestionCriteria,
			v components.False,
		) components.DecisionsNoulQuestionCriteria {
			c.False = v
			return c
		},
	)

	noulSpecKeyLens = L.MakeLens(
		func(s noulLabelSpec) string { return s.key },
		func(s noulLabelSpec, v string) noulLabelSpec {
			s.key = v
			return s
		},
	)

	noulSpecLabelLens = L.MakeLens(
		func(s noulLabelSpec) string { return s.label },
		func(s noulLabelSpec, v string) noulLabelSpec {
			s.label = v
			return s
		},
	)

	ansChoiceLens = L.MakeLens(
		func(a components.DecisionsResponseAnswers) *components.DecisionsChoiceAnswer {
			return a.DecisionsChoiceAnswer
		},
		func(
			a components.DecisionsResponseAnswers,
			v *components.DecisionsChoiceAnswer,
		) components.DecisionsResponseAnswers {
			a.DecisionsChoiceAnswer = v
			return a
		},
	)

	ansNoulLens = L.MakeLens(
		func(a components.DecisionsResponseAnswers) *components.DecisionsNoulAnswer {
			return a.DecisionsNoulAnswer
		},
		func(
			a components.DecisionsResponseAnswers,
			v *components.DecisionsNoulAnswer,
		) components.DecisionsResponseAnswers {
			a.DecisionsNoulAnswer = v
			return a
		},
	)

	isNonNilAllData     = P.IsNonZero[*allData]()
	isNonNilStrainData  = P.IsNonZero[*strainData]()
	isNonNilPlasmidData = P.IsNonZero[*plasmidData]()
	isNonNilLogger      = P.IsNonZero[*logrus.Entry]()
	isNonNilSDK         = P.IsNonZero[*openrouter.OpenRouter]()
	isNonNilAlpha       = P.IsNonZero[*openrouter.Alpha]()
	isNonNilDecisionsP  = P.IsNonZero[*openrouter.Decisions]()
	isNonNilResp        = P.IsNonZero[*components.DecisionsResponse]()
	isNonNilChoiceAns   = P.IsNonZero[*components.DecisionsChoiceAnswer]()
	isNonNilNoulAns     = P.IsNonZero[*components.DecisionsNoulAnswer]()
	isNonNilFloatPtr    = P.IsNonZero[*float64]()

	meetsThreshold = P.Not(N.LessThan(jevConfidenceThreshold))

	isValidChoiceStr = F.Pipe1(
		S.IsNonEmpty,
		P.And(predord.NotEqualStr(choiceNone)),
	)

	containsAxenic = F.Flow2(strings.ToLower, S.Includes("axenic"))
	containsGWDI   = F.Flow2(strings.ToUpper, S.Includes("GWDI"))
	containsBact   = F.Flow2(
		strings.ToLower,
		F.Pipe2(
			S.Includes("bacterial stock"),
			P.Or(S.Includes("k. aerogenes")),
			P.Or(S.Includes("klebsiella")),
		),
	)
	containsGoldenBraid = F.Flow2(
		strings.ToLower,
		F.Pipe1(
			S.Includes("pdgb"),
			P.Or(S.Includes("goldenbraid")),
		),
	)

	rowHasAxenic      = A.Any(containsAxenic)
	rowHasGWDI        = A.Any(containsGWDI)
	rowHasBact        = A.Any(containsBact)
	rowHasGoldenBraid = A.Any(containsGoldenBraid)

	rowsHaveAxenic      = A.Any(rowHasAxenic)
	rowsHaveGWDI        = A.Any(rowHasGWDI)
	rowsHaveBact        = A.Any(rowHasBact)
	rowsHaveGoldenBraid = A.Any(rowHasGoldenBraid)

	hasClientState = F.Pipe1(isNonNilDecisions, P.ContraMap(resolveClientLens.Get))
	hasLoggerState = F.Pipe1(isNonNilLogger, P.ContraMap(resolveLoggerLens.Get))

	baseFromState = F.Flow2(resolveBaseLens.Get, IOE.Of[error, []string])

	baseOrJevFold = P.Fold(baseFromState, jevLabelPipeline)

	jevOrBaseBranch = F.Pipe1(hasClientState, baseOrJevFold)

	choiceQuestionKeys = []string{
		questionInstitutionType,
		questionPaymentLabel,
		questionLocationLabel,
	}

	noulQuestionSpecs = []noulLabelSpec{
		F.Pipe2(
			noulLabelSpec{},
			noulSpecKeyLens.Set(questionIsMSI),
			noulSpecLabelLens.Set(labelMSI),
		),
		F.Pipe2(
			noulLabelSpec{},
			noulSpecKeyLens.Set(questionBacterialFood),
			noulSpecLabelLens.Set(labelBacterialFood),
		),
		F.Pipe2(
			noulLabelSpec{},
			noulSpecKeyLens.Set(questionGoldenBraid),
			noulSpecLabelLens.Set(labelGoldenBraid),
		),
		F.Pipe2(
			noulLabelSpec{},
			noulSpecKeyLens.Set(questionGWDI),
			noulSpecLabelLens.Set(labelGWDI),
		),
		F.Pipe2(
			noulLabelSpec{},
			noulSpecKeyLens.Set(questionAxenicMedium),
			noulSpecLabelLens.Set(labelAxenicMedium),
		),
	}
)

func sdkAlpha(sdk *openrouter.OpenRouter) O.Option[*openrouter.Alpha] {
	return F.Pipe1(sdk.Alpha, O.FromPredicate(isNonNilAlpha))
}

func alphaDecisions(alpha *openrouter.Alpha) O.Option[DecisionsClient] {
	return F.Pipe2(
		alpha.Decisions,
		O.FromPredicate(isNonNilDecisionsP),
		O.Map(func(d *openrouter.Decisions) DecisionsClient { return d }),
	)
}

func resolveJevModel(model string) string {
	return F.Pipe3(
		model,
		strings.TrimSpace,
		O.FromPredicate(S.IsNonEmpty),
		O.GetOrElse(F.Constant(defaultJevModel)),
	)
}

func (gh *githubIssue) resolveLabels(all *allData, ord *order.Order) []string {
	base := deterministicLabels(all)

	return F.Pipe9(
		labelResolveState{},
		resolveClientLens.Set(ghDecisionsLens.Get(*gh)),
		resolveLoggerLens.Set(ghLoggerLens.Get(*gh)),
		resolveModelLens.Set(resolveJevModel(ghJevModelLens.Get(*gh))),
		resolveDataLens.Set(all),
		resolveOrderLens.Set(ord),
		resolveBaseLens.Set(base),
		jevOrBaseBranch,
		ioeutils.ToEither[error, []string],
		E.GetOrElse(F.Constant1[error](base)),
	)
}

func buildStateRequest(s labelResolveState) components.DecisionsRequest {
	stateMap := buildDecisionsState(
		resolveDataLens.Get(s),
		resolveOrderLens.Get(s),
	)

	return F.Pipe3(
		components.DecisionsRequest{},
		decReqModelLens.Set(resolveModelLens.Get(s)),
		decReqStateLens.Set(components.CreateStateMapOfAny(stateMap)),
		decReqQuestionsLens.Set(buildDecisionsQuestions()),
	)
}

func executeDecisionsCall(
	s labelResolveState,
) IOE.IOEither[error, *components.DecisionsResponse] {
	client := resolveClientLens.Get(s)
	req := resolveReqLens.Get(s)

	return IOE.TryCatchError(func() (*components.DecisionsResponse, error) {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			jevRequestTimeout,
		)
		defer cancel()

		return client.Create(ctx, req)
	})
}

func logJevWarning(s labelResolveState, err error) IO.IO[IO.Void] {
	warn := func(st labelResolveState) IO.IO[IO.Void] {
		return IO.FromImpure(func() {
			resolveLoggerLens.Get(st).Warnf(
				"openrouter jev label classification failed: %v",
				err,
			)
		})
	}

	return F.Pipe2(
		s,
		O.FromPredicate(hasLoggerState),
		O.Fold(F.Constant(IO.Of(IO.Void{})), warn),
	)
}

func fallbackBaseLabels(
	s labelResolveState,
) func(error) IOE.IOEither[error, []string] {
	base := resolveBaseLens.Get(s)

	return func(err error) IOE.IOEither[error, []string] {
		return F.Pipe2(
			logJevWarning(s, err),
			IOE.FromIO[error],
			IOE.MapTo[error, IO.Void](base),
		)
	}
}

func jevLabelPipeline(
	s labelResolveState,
) IOE.IOEither[error, []string] {
	return F.Pipe5(
		s,
		IOE.Of[error],
		IOE.Let[error](resolveReqLens.Set, buildStateRequest),
		IOE.Bind(resolveRespLens.Set, executeDecisionsCall),
		IOE.Map[error](func(st labelResolveState) []string {
			return F.Pipe6(
				st,
				resolveRespLens.Get,
				O.FromPredicate(isNonNilResp),
				O.Map(
					func(r *components.DecisionsResponse) map[string]components.DecisionsResponseAnswers {
						return r.GetAnswers()
					},
				),
				O.Fold(
					F.Constant([]string(nil)),
					extractAnswersLabels,
				),
				func(jevLabels []string) []string {
					return A.ArrayConcatAll(resolveBaseLens.Get(st), jevLabels)
				},
				deduplicateLabels,
			)
		}),
		IOE.OrElse(fallbackBaseLabels(s)),
	)
}

func optStrainData(all *allData) O.Option[*strainData] {
	return F.Pipe3(
		all,
		O.FromPredicate(isNonNilAllData),
		O.Map(func(d *allData) *strainData { return allStrainDataLens.Get(*d) }),
		O.Chain(O.FromPredicate(isNonNilStrainData)),
	)
}

func optPlasmidData(all *allData) O.Option[*plasmidData] {
	return F.Pipe3(
		all,
		O.FromPredicate(isNonNilAllData),
		O.Map(func(d *allData) *plasmidData { return allPlasmidDataLens.Get(*d) }),
		O.Chain(O.FromPredicate(isNonNilPlasmidData)),
	)
}

func strainRowsCombined(sd *strainData) [][]string {
	return A.ArrayConcatAll(
		strainDataInfoLens.Get(*sd),
		strainDataInvsLens.Get(*sd),
	)
}

func hasStrains(all *allData) bool {
	return F.Pipe3(
		all,
		optStrainData,
		O.Map(func(sd *strainData) []*stock.Strain {
			return strainDataItemsLens.Get(*sd)
		}),
		O.Fold(F.Constant(false), predarrays.IsNonEmpty[*stock.Strain]()),
	)
}

func hasPlasmids(all *allData) bool {
	return F.Pipe3(
		all,
		optPlasmidData,
		O.Map(func(pd *plasmidData) []*stock.Plasmid {
			return plasmidDataItemsLens.Get(*pd)
		}),
		O.Fold(F.Constant(false), predarrays.IsNonEmpty[*stock.Plasmid]()),
	)
}

func hasAxenicStorage(all *allData) bool {
	return F.Pipe3(
		all,
		optStrainData,
		O.Map(func(sd *strainData) [][]string {
			return strainDataInvsLens.Get(*sd)
		}),
		O.Fold(F.Constant(false), rowsHaveAxenic),
	)
}

func hasGWDIStrains(all *allData) bool {
	return F.Pipe3(
		all,
		optStrainData,
		O.Map(strainRowsCombined),
		O.Fold(F.Constant(false), rowsHaveGWDI),
	)
}

func hasBacterialFoodSource(all *allData) bool {
	return F.Pipe3(
		all,
		optStrainData,
		O.Map(strainRowsCombined),
		O.Fold(F.Constant(false), rowsHaveBact),
	)
}

func hasGoldenBraidPlasmids(all *allData) bool {
	return F.Pipe3(
		all,
		optPlasmidData,
		O.Map(func(pd *plasmidData) [][]string {
			return plasmidDataInvsLens.Get(*pd)
		}),
		O.Fold(F.Constant(false), rowsHaveGoldenBraid),
	)
}

func labelIf(pred func(*allData) bool, label string) func(*allData) O.Option[string] {
	return F.Flow2(
		O.FromPredicate(pred),
		O.Map(F.Constant1[*allData](label)),
	)
}

func deterministicLabels(all *allData) []string {
	return F.Pipe2(
		[]O.Option[string]{
			F.Pipe1(all, labelIf(hasStrains, labelStrainOrder)),
			F.Pipe1(all, labelIf(hasPlasmids, labelPlasmidOrder)),
			F.Pipe1(all, labelIf(hasAxenicStorage, labelAxenicMedium)),
			F.Pipe1(all, labelIf(hasGWDIStrains, labelGWDI)),
			F.Pipe1(all, labelIf(hasBacterialFoodSource, labelBacterialFood)),
			F.Pipe1(all, labelIf(hasGoldenBraidPlasmids, labelGoldenBraid)),
		},
		arrutils.Compact[string],
		deduplicateLabels,
	)
}

func extractStrainInfoRows(all *allData) [][]string {
	return F.Pipe3(
		all,
		optStrainData,
		O.Map(func(sd *strainData) [][]string {
			return strainDataInfoLens.Get(*sd)
		}),
		O.GetOrElse(F.Constant([][]string(nil))),
	)
}

func extractStrainInvRows(all *allData) [][]string {
	return F.Pipe3(
		all,
		optStrainData,
		O.Map(func(sd *strainData) [][]string {
			return strainDataInvsLens.Get(*sd)
		}),
		O.GetOrElse(F.Constant([][]string(nil))),
	)
}

func extractPlasmidInvRows(all *allData) [][]string {
	return F.Pipe3(
		all,
		optPlasmidData,
		O.Map(func(pd *plasmidData) [][]string {
			return plasmidDataInvsLens.Get(*pd)
		}),
		O.GetOrElse(F.Constant([][]string(nil))),
	)
}

func buildDecisionsState(all *allData, ord *order.Order) map[string]any {
	attr := ord.GetData().GetAttributes()
	shipper := attr.GetConsumerInfo()
	payer := attr.GetPayerInfo()

	return map[string]any{
		"order_id":           ord.GetData().GetId(),
		"purchaser":          attr.GetPurchaser(),
		"consumer_email":     attr.GetConsumer(),
		"payer_email":        attr.GetPayer(),
		"payment_method":     attr.GetPayment(),
		"purchase_order_num": attr.GetPurchaseOrderNum(),
		"courier":            attr.GetCourier(),
		"comments":           attr.GetComments(),
		"items":              attr.GetItems(),
		"shipping_address": map[string]string{
			"first_name":     shipper.GetFirstName(),
			"last_name":      shipper.GetLastName(),
			"organization":   shipper.GetOrganization(),
			"first_address":  shipper.GetFirstAddress(),
			"second_address": shipper.GetSecondAddress(),
			"city":           shipper.GetCity(),
			"state":          shipper.GetState(),
			"zipcode":        shipper.GetZipcode(),
			"country":        shipper.GetCountry(),
		},
		"billing_address": map[string]string{
			"organization": payer.GetOrganization(),
			"city":         payer.GetCity(),
			"state":        payer.GetState(),
			"country":      payer.GetCountry(),
		},
		"strain_info":       extractStrainInfoRows(all),
		"strain_inventory":  extractStrainInvRows(all),
		"plasmid_inventory": extractPlasmidInvRows(all),
	}
}

func buildDecisionsQuestions() map[string]components.Questions {
	return map[string]components.Questions{
		questionInstitutionType: choiceQuestion(
			"Classify the customer's shipping organization into the Stock-Center-Orders institution category.",
			institutionCriteriaMap(),
		),
		questionIsMSI: noulQuestion(
			"Is the customer's US institution a Minority Serving Institution (HBCU, HSI, AANAPISI, or Tribal College)?",
			"The organization is a US Minority Serving Institution.",
			"The organization is not a US Minority Serving Institution or is outside the US.",
		),
		questionPaymentLabel: choiceQuestion(
			"Select the Stock-Center-Orders payment method label matching the order's payment method, PO number, and comments.",
			paymentCriteriaMap(),
		),
		questionLocationLabel: choiceQuestion(
			"Select the geographic label for the shipping address: 2-letter US state code or international country label.",
			locationCriteriaMap(),
		),
		questionBacterialFood: noulQuestion(
			"Does the order include bacterial food source strains (Klebsiella aerogenes, E. coli B/r, or bacterial stock)?",
			"Order includes bacterial food source or bacterial stock.",
			"Order does not include bacterial food source.",
		),
		questionGoldenBraid: noulQuestion(
			"Does the order include GoldenBraid plasmids (such as pDGB plasmids or GoldenBraid mentioned in items/comments)?",
			"Order includes GoldenBraid plasmids.",
			"Order does not include GoldenBraid plasmids.",
		),
		questionGWDI: noulQuestion(
			"Does the order include Genome-Wide Dictyostelium Insertion (GWDI) mutant strains?",
			"Order includes GWDI strains.",
			"Order does not include GWDI strains.",
		),
		questionAxenicMedium: noulQuestion(
			"Does the order include strains stored or grown in axenic medium?",
			"Order includes axenic strains or axenic medium.",
			"Order does not include axenic medium.",
		),
	}
}

func institutionCriteriaMap() map[string]string {
	return map[string]string{
		labelResearchInst:     "University, research institute, medical school, or academic research lab.",
		labelUndergradCollege: "Primarily undergraduate liberal arts college or community college.",
		labelHighSchool:       "High school, secondary school, preparatory academy, or K-12 school.",
		labelCommercial:       "For-profit biotechnology, pharmaceutical, or corporate company.",
		choiceNone:            "Cannot determine institution type from the address.",
	}
}

func paymentCriteriaMap() map[string]string {
	return R.FromEntries([]PR.Pair[string, string]{
		PR.MakePair(labelCreditCard, "Credit card or procard payment."),
		PR.MakePair(labelPurchaseOrder, "Purchase order (PO) payment."),
		PR.MakePair(labelWireTransfer, "Wire transfer or bank transfer payment."),
		PR.MakePair(
			labelChartString,
			"Northwestern University internal chart string payment.",
		),
		PR.MakePair(labelWaiverNoPay, "Fee waiver requested or no payment."),
		PR.MakePair(
			labelNoCharge,
			"Free of charge or replacement order at no charge.",
		),
		PR.MakePair(choiceNone, "Payment method is unspecified."),
	})
}

func toCriteriaPtr(desc string) *components.Criteria {
	return new(components.CreateCriteriaStr(desc))
}

func choiceQuestion(
	instructions string,
	options map[string]string,
) components.Questions {
	inst := components.CreateDecisionsChoiceQuestionInstructionsStr(instructions)

	return F.Pipe4(
		components.DecisionsChoiceQuestion{},
		choiceQCriteriaLens.Set(F.Pipe1(options, R.Map[string](toCriteriaPtr))),
		choiceQInstructionsLens.Set(inst),
		choiceQTypeLens.Set(components.DecisionsChoiceQuestionTypeChoice),
		components.CreateQuestionsChoice,
	)
}

func noulQuestion(
	instructions string,
	trueDesc string,
	falseDesc string,
) components.Questions {
	crit := F.Pipe2(
		components.DecisionsNoulQuestionCriteria{},
		noulCriteriaTrueLens.Set(components.CreateTrueStr(trueDesc)),
		noulCriteriaFalseLens.Set(components.CreateFalseStr(falseDesc)),
	)
	inst := components.CreateDecisionsNoulQuestionInstructionsStr(instructions)

	return F.Pipe4(
		components.DecisionsNoulQuestion{},
		noulQCriteriaLens.Set(new(crit)),
		noulQInstructionsLens.Set(inst),
		noulQTypeLens.Set(components.DecisionsNoulQuestionTypeNoul),
		components.CreateQuestionsNoul,
	)
}

func choiceConfidenceScore(ans *components.DecisionsChoiceAnswer) float64 {
	choice := strings.TrimSpace(ans.GetChoice())
	probs := ans.GetProbabilities()

	return F.Pipe3(
		ans.GetConfidence(),
		O.FromPredicate(isNonNilFloatPtr),
		O.Map(func(v *float64) float64 { return *v }),
		O.GetOrElse(func() float64 {
			return F.Pipe2(
				probs,
				R.Lookup[float64](choice),
				O.GetOrElse(F.Constant(0.0)),
			)
		}),
	)
}

func isConfidentChoiceAns(ans *components.DecisionsChoiceAnswer) bool {
	return F.Pipe2(
		ans,
		choiceConfidenceScore,
		meetsThreshold,
	)
}

func extractChoiceOption(
	ans components.DecisionsResponseAnswers,
) O.Option[string] {
	return F.Pipe5(
		ans,
		ansChoiceLens.Get,
		O.FromPredicate(isNonNilChoiceAns),
		O.Chain(O.FromPredicate(isConfidentChoiceAns)),
		O.Map(func(a *components.DecisionsChoiceAnswer) string {
			return strings.TrimSpace(a.GetChoice())
		}),
		O.Chain(O.FromPredicate(isValidChoiceStr)),
	)
}

func lookupChoiceLabel(
	answers map[string]components.DecisionsResponseAnswers,
	key string,
) O.Option[string] {
	return F.Pipe2(
		answers,
		R.Lookup[components.DecisionsResponseAnswers](key),
		O.Chain(extractChoiceOption),
	)
}

func isHighConfidenceNoul(ans components.DecisionsResponseAnswers) bool {
	return F.Pipe3(
		ans,
		ansNoulLens.Get,
		O.FromPredicate(isNonNilNoulAns),
		O.Fold(
			F.Constant(false),
			F.Flow2(
				func(a *components.DecisionsNoulAnswer) float64 {
					return a.GetNoul()
				},
				meetsThreshold,
			),
		),
	)
}

func lookupNoulLabel(
	answers map[string]components.DecisionsResponseAnswers,
	spec noulLabelSpec,
) O.Option[string] {
	return F.Pipe3(
		answers,
		R.Lookup[components.DecisionsResponseAnswers](noulSpecKeyLens.Get(spec)),
		O.Chain(O.FromPredicate(isHighConfidenceNoul)),
		O.Map(
			F.Constant1[components.DecisionsResponseAnswers](
				noulSpecLabelLens.Get(spec),
			),
		),
	)
}

func extractAnswersLabels(
	answers map[string]components.DecisionsResponseAnswers,
) []string {
	choiceLabels := F.Pipe2(
		choiceQuestionKeys,
		A.Map(F.Bind1st(lookupChoiceLabel, answers)),
		arrutils.Compact[string],
	)
	noulLabels := F.Pipe2(
		noulQuestionSpecs,
		A.Map(F.Bind1st(lookupNoulLabel, answers)),
		arrutils.Compact[string],
	)

	return A.ArrayConcatAll(choiceLabels, noulLabels)
}

func deduplicateLabels(labels []string) []string {
	seen := mapset.NewThreadUnsafeSet[string]()
	isFreshLabel := func(label string) bool {
		return seen.Add(label)
	}

	return F.Pipe3(
		labels,
		A.Map(strings.TrimSpace),
		A.Filter(S.IsNonEmpty),
		A.Filter(isFreshLabel),
	)
}

func formatUSStateDescription(entry PR.Pair[string, string]) PR.Pair[string, string] {
	code := PR.Head(entry)
	stateName := PR.Tail(entry)
	desc := strutils.JoinStrings(
		"US shipping address in ",
		stateName,
		" ",
		strutils.Parens(code),
		".",
	)

	return PR.MakePair(code, desc)
}

func locationCriteriaMap() map[string]string {
	intl := map[string]string{
		"Washington D.C":                        "US shipping address in Washington, D.C.",
		"Canada":                                "Shipping address in Canada.",
		"Germany":                               "Shipping address in Germany.",
		"France":                                "Shipping address in France.",
		"Italy":                                 "Shipping address in Italy.",
		"United Kingdom (UK)":                   "Shipping address in the United Kingdom (England, Wales, Northern Ireland).",
		"Scotland":                              "Shipping address in Scotland.",
		"India":                                 "Shipping address in India.",
		"Korea Democratic People's Republic of": "Shipping address in South Korea / Republic of Korea.",
		"Brazil":                                "Shipping address in Brazil.",
		"Japan":                                 "Shipping address in Japan.",
		"Netherlands":                           "Shipping address in the Netherlands.",
		"Spain":                                 "Shipping address in Spain.",
		"Chile":                                 "Shipping address in Chile.",
		"Switzerland":                           "Shipping address in Switzerland.",
		"Austria":                               "Shipping address in Austria.",
		"Australia":                             "Shipping address in Australia.",
		"Taiwan Province of China":              "Shipping address in Taiwan.",
		"Sweden":                                "Shipping address in Sweden.",
		"Singapore":                             "Shipping address in Singapore.",
		"Argentina":                             "Shipping address in Argentina.",
		"China":                                 "Shipping address in China.",
		"Denmark":                               "Shipping address in Denmark.",
		"Israel":                                "Shipping address in Israel.",
		"Costa Rica":                            "Shipping address in Costa Rica.",
		"Turkey":                                "Shipping address in Turkey.",
		"Mexico":                                "Shipping address in Mexico.",
		"New Zealand":                           "Shipping address in New Zealand.",
		"Thailand":                              "Shipping address in Thailand.",
		"Saudi Arabia":                          "Shipping address in Saudi Arabia.",
		"Portugal":                              "Shipping address in Portugal.",
		"Croatia":                               "Shipping address in Croatia.",
		"UAE":                                   "Shipping address in the United Arab Emirates.",
		"Czech Republic":                        "Shipping address in the Czech Republic.",
		choiceNone:                              "Location does not match any listed US state or country.",
	}

	states := F.Pipe3(
		usStateLabels(),
		R.ToEntries[string, string],
		A.Map(formatUSStateDescription),
		R.FromEntries[string, string],
	)

	return F.Pipe1(
		[]map[string]string{states, intl},
		M.ConcatAll(R.UnionLastMonoid[string, string]()),
	)
}

func usStateLabels() map[string]string {
	return map[string]string{
		"AL": "Alabama", "AK": "Alaska", "AZ": "Arizona", "AR": "Arkansas",
		"CA": "California", "CO": "Colorado", "CT": "Connecticut", "DE": "Delaware",
		"FL": "Florida", "GA": "Georgia", "HI": "Hawaii", "ID": "Idaho",
		"IL": "Illinois", "IN": "Indiana", "IA": "Iowa", "KS": "Kansas",
		"KY": "Kentucky", "LA": "Louisiana", "ME": "Maine", "MD": "Maryland",
		"MA": "Massachusetts", "MI": "Michigan", "MN": "Minnesota", "MS": "Mississippi",
		"MO": "Missouri", "MT": "Montana", "NE": "Nebraska", "NV": "Nevada",
		"NH": "New Hampshire", "NJ": "New Jersey", "NM": "New Mexico", "NY": "New York",
		"NC": "North Carolina", "ND": "North Dakota", "OH": "Ohio", "OK": "Oklahoma",
		"OR": "Oregon", "PA": "Pennsylvania", "RI": "Rhode Island", "SC": "South Carolina",
		"SD": "South Dakota", "TN": "Tennessee", "TX": "Texas", "UT": "Utah",
		"VT": "Vermont", "VA": "Virginia", "WA": "Washington", "WV": "West Virginia",
		"WI": "Wisconsin", "WY": "Wyoming",
	}
}
