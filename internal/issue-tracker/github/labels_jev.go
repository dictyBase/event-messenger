package github

import (
	"context"
	"strings"
	"time"

	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	IOE "github.com/IBM/fp-go/v2/ioeither"
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
	arrutils "github.com/dictyBase/fp-go-loom/array"
	MO "github.com/dictyBase/fp-go-loom/matchopt"
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

type answersMap = map[string]components.DecisionsResponseAnswers

type questionsMap = map[string]components.Questions

type labelResolveState struct {
	Client  DecisionsClient
	Logger  *logrus.Entry
	Model   string
	Data    *allData
	Order   *order.Order
	Base    []string
	Request components.DecisionsRequest
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
		func(r components.DecisionsRequest) questionsMap { return r.Questions },
		func(
			r components.DecisionsRequest,
			v questionsMap,
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

	noulCriteriaTrueLens = L.MakeLensRef(
		func(c *components.DecisionsNoulQuestionCriteria) components.True { return c.True },
		func(
			c *components.DecisionsNoulQuestionCriteria,
			v components.True,
		) *components.DecisionsNoulQuestionCriteria {
			c.True = v
			return c
		},
	)

	noulCriteriaFalseLens = L.MakeLensRef(
		func(c *components.DecisionsNoulQuestionCriteria) components.False { return c.False },
		func(
			c *components.DecisionsNoulQuestionCriteria,
			v components.False,
		) *components.DecisionsNoulQuestionCriteria {
			c.False = v
			return c
		},
	)

	noulQTrueLens = F.Pipe1(
		noulQCriteriaLens,
		L.Compose[components.DecisionsNoulQuestion](noulCriteriaTrueLens),
	)

	noulQFalseLens = F.Pipe1(
		noulQCriteriaLens,
		L.Compose[components.DecisionsNoulQuestion](noulCriteriaFalseLens),
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

	sdkAlphaLens = L.MakeLensRef(
		func(s *openrouter.OpenRouter) *openrouter.Alpha { return s.Alpha },
		func(s *openrouter.OpenRouter, v *openrouter.Alpha) *openrouter.OpenRouter {
			s.Alpha = v
			return s
		},
	)

	alphaDecisionsLens = L.MakeLensRef(
		func(a *openrouter.Alpha) *openrouter.Decisions { return a.Decisions },
		func(a *openrouter.Alpha, v *openrouter.Decisions) *openrouter.Alpha {
			a.Decisions = v
			return a
		},
	)

	isNonNilLogger     = P.IsNonZero[*logrus.Entry]()
	isNonNilSDK        = P.IsNonZero[*openrouter.OpenRouter]()
	isNonNilAlpha      = P.IsNonZero[*openrouter.Alpha]()
	isNonNilDecisionsP = P.IsNonZero[*openrouter.Decisions]()
	isNonNilResp       = P.IsNonZero[*components.DecisionsResponse]()
	isNonNilChoiceAns  = P.IsNonZero[*components.DecisionsChoiceAnswer]()
	isNonNilNoulAns    = P.IsNonZero[*components.DecisionsNoulAnswer]()

	meetsThreshold = P.Not(N.LessThan(jevConfidenceThreshold))
	notNoneChoice  = predord.NotEqualStr(choiceNone)

	isValidChoiceStr = F.Pipe1(
		S.IsNonEmpty,
		P.And(notNoneChoice),
	)

	containsAxenic = F.Pipe1(S.Includes("axenic"), P.ContraMap(strings.ToLower))
	containsGWDI   = F.Pipe1(S.Includes("GWDI"), P.ContraMap(strings.ToUpper))

	hasBacterialTerm = F.Pipe2(
		S.Includes("bacterial stock"),
		P.Or(S.Includes("k. aerogenes")),
		P.Or(S.Includes("klebsiella")),
	)
	containsBact = F.Pipe1(hasBacterialTerm, P.ContraMap(strings.ToLower))

	hasGoldenBraidTerm = F.Pipe1(
		S.Includes("pdgb"),
		P.Or(S.Includes("goldenbraid")),
	)
	containsGoldenBraid = F.Pipe1(
		hasGoldenBraidTerm,
		P.ContraMap(strings.ToLower),
	)

	rowHasAxenic      = A.Any(containsAxenic)
	rowHasGWDI        = A.Any(containsGWDI)
	rowHasBact        = A.Any(containsBact)
	rowHasGoldenBraid = A.Any(containsGoldenBraid)

	rowsHaveAxenic      = A.Any(rowHasAxenic)
	rowsHaveGWDI        = A.Any(rowHasGWDI)
	rowsHaveBact        = A.Any(rowHasBact)
	rowsHaveGoldenBraid = A.Any(rowHasGoldenBraid)

	prependLabels  = F.Flip(A.Concat[string])
	mapCriteriaPtr = R.Map[string](
		F.Flow2(components.CreateCriteriaStr, F.Ref[components.Criteria]),
	)

	choiceConfidenceScore = MO.Match(
		0.0,
		[]O.Kleisli[*components.DecisionsChoiceAnswer, float64]{
			F.Flow2(
				(*components.DecisionsChoiceAnswer).GetConfidence,
				O.FromNillable2[float64],
			),
			choiceProbabilityOpt,
		},
	)

	isConfidentChoiceAns = F.Pipe1(
		meetsThreshold,
		P.ContraMap(choiceConfidenceScore),
	)

	deterministicLabelRules = []func(*allData) O.Option[string]{
		MO.Const(hasStrains, labelStrainOrder),
		MO.Const(hasPlasmids, labelPlasmidOrder),
		MO.Const(hasAxenicStorage, labelAxenicMedium),
		MO.Const(hasGWDIStrains, labelGWDI),
		MO.Const(hasBacterialFoodSource, labelBacterialFood),
		MO.Const(hasGoldenBraidPlasmids, labelGoldenBraid),
	}

	jevAnswerExtractors = []func(answersMap) O.Option[string]{
		choiceExtractor(questionInstitutionType),
		choiceExtractor(questionPaymentLabel),
		choiceExtractor(questionLocationLabel),
		noulExtractor(questionIsMSI, labelMSI),
		noulExtractor(questionBacterialFood, labelBacterialFood),
		noulExtractor(questionGoldenBraid, labelGoldenBraid),
		noulExtractor(questionGWDI, labelGWDI),
		noulExtractor(questionAxenicMedium, labelAxenicMedium),
	}

	resolveParamsDecisions = MO.Match(
		DecisionsClient(nil),
		[]O.Kleisli[*IssueParams, DecisionsClient]{
			F.Flow2(
				IssueParamsDecisionsLens.Get,
				O.FromPredicate(isNonNilDecisions),
			),
			F.Flow12(
				IssueParamsOpenRouterLens.Get,
				strings.TrimSpace,
				O.FromPredicate(S.IsNonEmpty),
				O.Map(openrouter.WithSecurity),
				O.Map(A.Of[openrouter.SDKOption]),
				O.Map(F.Unvariadic0(openrouter.New)),
				O.Chain(O.FromPredicate(isNonNilSDK)),
				O.Map(sdkAlphaLens.Get),
				O.Chain(O.FromPredicate(isNonNilAlpha)),
				O.Map(alphaDecisionsLens.Get),
				O.Chain(O.FromPredicate(isNonNilDecisionsP)),
				O.Map(toDecisionsClient),
			),
		},
	)
)

func toDecisionsClient(d *openrouter.Decisions) DecisionsClient {
	return d
}

func resolveJevModel(model string) string {
	return F.Pipe3(
		model,
		strings.TrimSpace,
		O.FromPredicate(S.IsNonEmpty),
		O.GetOrElse(F.Constant(defaultJevModel)),
	)
}

func jevOrBaseBranch(s labelResolveState) IOE.IOEither[error, []string] {
	return F.Pipe3(
		isNonNilDecisions,
		P.ContraMap(resolveClientLens.Get),
		P.Fold(
			F.Flow2(resolveBaseLens.Get, IOE.Of[error, []string]),
			jevLabelPipeline,
		),
		F.Bind1st(
			F.Pipe1[func(labelResolveState) IOE.IOEither[error, []string]],
			s,
		),
	)
}

func (gh *githubIssue) resolveLabels(
	s issueCreateState,
) IOE.IOEither[error, []string] {
	return F.Pipe8(
		labelResolveState{},
		F.Pipe2(*gh, ghDecisionsLens.Get, resolveClientLens.Set),
		F.Pipe2(*gh, ghLoggerLens.Get, resolveLoggerLens.Set),
		F.Pipe3(
			*gh,
			ghJevModelLens.Get,
			resolveJevModel,
			resolveModelLens.Set,
		),
		F.Pipe2(s, createStateDataLens.Get, resolveDataLens.Set),
		F.Pipe2(s, createStateOrderLens.Get, resolveOrderLens.Set),
		IOE.Of[error],
		IOE.Let[error](
			resolveBaseLens.Set,
			F.Flow2(resolveDataLens.Get, deterministicLabels),
		),
		IOE.Chain(jevOrBaseBranch),
	)
}

func buildStateRequest(s labelResolveState) components.DecisionsRequest {
	return F.Pipe4(
		buildDecisionsState(
			resolveDataLens.Get(s),
			resolveOrderLens.Get(s),
		),
		components.CreateStateMapOfAny,
		F.Pipe1(components.DecisionsRequest{}, F.Flip(decReqStateLens.Set)),
		F.Pipe2(s, resolveModelLens.Get, decReqModelLens.Set),
		F.Pipe2(
			s,
			F.Ignore1of1[labelResolveState](buildDecisionsQuestions),
			decReqQuestionsLens.Set,
		),
	)
}

func executeDecisionsCall(
	s labelResolveState,
) IOE.IOEither[error, *components.DecisionsResponse] {
	return IOE.TryCatchError(func() (*components.DecisionsResponse, error) {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			jevRequestTimeout,
		)
		defer cancel()

		return resolveClientLens.Get(s).Create(ctx, resolveReqLens.Get(s))
	})
}

func emitJevWarning(err error, lg *logrus.Entry) IO.IO[IO.Void] {
	return IO.FromImpure(func() {
		lg.Warnf("openrouter jev label classification failed: %v", err)
	})
}

func logJevWarning(err error, s labelResolveState) IO.IO[IO.Void] {
	return F.Pipe3(
		s,
		resolveLoggerLens.Get,
		O.FromPredicate(isNonNilLogger),
		O.Fold(
			F.Constant(IO.Of(IO.Void{})),
			F.Bind1st(emitJevWarning, err),
		),
	)
}

func recoverBaseLabels(
	s labelResolveState,
	err error,
) IOE.IOEither[error, []string] {
	return F.Pipe3(
		s,
		F.Bind1st(logJevWarning, err),
		IOE.FromIO[error],
		F.Pipe2(s, resolveBaseLens.Get, IOE.MapTo[error, IO.Void]),
	)
}

func jevLabelPipeline(
	s labelResolveState,
) IOE.IOEither[error, []string] {
	return F.Pipe9(
		s,
		IOE.Of[error],
		IOE.Let[error](resolveReqLens.Set, buildStateRequest),
		IOE.Chain(executeDecisionsCall),
		IOE.Map[error](O.FromPredicate(isNonNilResp)),
		IOE.Map[error](O.Map((*components.DecisionsResponse).GetAnswers)),
		IOE.Map[error](O.Fold(F.Constant([]string(nil)), extractAnswersLabels)),
		IOE.Map[error](F.Pipe2(s, resolveBaseLens.Get, prependLabels)),
		IOE.Map[error](deduplicateLabels),
		IOE.OrElse(F.Bind1st(recoverBaseLabels, s)),
	)
}

func optStrainData(all *allData) O.Option[strainData] {
	return F.Pipe3(
		all,
		O.FromNillable2[allData],
		O.Map(allStrainDataLens.Get),
		O.Chain(O.FromNillable2[strainData]),
	)
}

func optPlasmidData(all *allData) O.Option[plasmidData] {
	return F.Pipe3(
		all,
		O.FromNillable2[allData],
		O.Map(allPlasmidDataLens.Get),
		O.Chain(O.FromNillable2[plasmidData]),
	)
}

func strainRowsCombined(sd strainData) [][]string {
	return A.ArrayConcatAll(
		strainDataInfoLens.Get(sd),
		strainDataInvsLens.Get(sd),
	)
}

func hasStrains(all *allData) bool {
	return F.Pipe3(
		all,
		optStrainData,
		O.Map(strainDataItemsLens.Get),
		O.Fold(F.Constant(false), predarrays.IsNonEmpty[*stock.Strain]()),
	)
}

func hasPlasmids(all *allData) bool {
	return F.Pipe3(
		all,
		optPlasmidData,
		O.Map(plasmidDataItemsLens.Get),
		O.Fold(F.Constant(false), predarrays.IsNonEmpty[*stock.Plasmid]()),
	)
}

func hasAxenicStorage(all *allData) bool {
	return F.Pipe3(
		all,
		optStrainData,
		O.Map(strainDataInvsLens.Get),
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
		O.Map(plasmidDataInvsLens.Get),
		O.Fold(F.Constant(false), rowsHaveGoldenBraid),
	)
}

func deterministicLabels(all *allData) []string {
	return F.Pipe3(
		deterministicLabelRules,
		A.Flap[O.Option[string]](all),
		arrutils.Compact[string],
		deduplicateLabels,
	)
}

func extractStrainInfoRows(all *allData) [][]string {
	return F.Pipe3(
		all,
		optStrainData,
		O.Map(strainDataInfoLens.Get),
		O.GetOrElse(F.Constant([][]string(nil))),
	)
}

func extractStrainInvRows(all *allData) [][]string {
	return F.Pipe3(
		all,
		optStrainData,
		O.Map(strainDataInvsLens.Get),
		O.GetOrElse(F.Constant([][]string(nil))),
	)
}

func extractPlasmidInvRows(all *allData) [][]string {
	return F.Pipe3(
		all,
		optPlasmidData,
		O.Map(plasmidDataInvsLens.Get),
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

func buildDecisionsQuestions() questionsMap {
	return questionsMap{
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

func choiceQuestion(
	instructions string,
	options map[string]string,
) components.Questions {
	return F.Pipe4(
		components.DecisionsChoiceQuestion{},
		F.Pipe2(options, mapCriteriaPtr, choiceQCriteriaLens.Set),
		F.Pipe2(
			instructions,
			components.CreateDecisionsChoiceQuestionInstructionsStr,
			choiceQInstructionsLens.Set,
		),
		choiceQTypeLens.Set(components.DecisionsChoiceQuestionTypeChoice),
		components.CreateQuestionsChoice,
	)
}

func noulQuestion(
	instructions string,
	trueDesc string,
	falseDesc string,
) components.Questions {
	return F.Pipe6(
		components.DecisionsNoulQuestion{},
		noulQCriteriaLens.Set(&components.DecisionsNoulQuestionCriteria{}),
		F.Pipe2(trueDesc, components.CreateTrueStr, noulQTrueLens.Set),
		F.Pipe2(falseDesc, components.CreateFalseStr, noulQFalseLens.Set),
		F.Pipe2(
			instructions,
			components.CreateDecisionsNoulQuestionInstructionsStr,
			noulQInstructionsLens.Set,
		),
		noulQTypeLens.Set(components.DecisionsNoulQuestionTypeNoul),
		components.CreateQuestionsNoul,
	)
}

func trimmedChoiceText(ans *components.DecisionsChoiceAnswer) string {
	return F.Pipe2(
		ans,
		(*components.DecisionsChoiceAnswer).GetChoice,
		strings.TrimSpace,
	)
}

func choiceProbabilityOpt(
	ans *components.DecisionsChoiceAnswer,
) O.Option[float64] {
	return F.Pipe2(
		ans,
		(*components.DecisionsChoiceAnswer).GetProbabilities,
		F.Pipe2(ans, trimmedChoiceText, R.Lookup[float64, string]),
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
		O.Map(trimmedChoiceText),
		O.Chain(O.FromPredicate(isValidChoiceStr)),
	)
}

func choiceExtractor(key string) func(answersMap) O.Option[string] {
	return F.Flow2(
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
			F.Flow2((*components.DecisionsNoulAnswer).GetNoul, meetsThreshold),
		),
	)
}

func noulExtractor(
	key string,
	label string,
) func(answersMap) O.Option[string] {
	return F.Flow3(
		R.Lookup[components.DecisionsResponseAnswers](key),
		O.Chain(O.FromPredicate(isHighConfidenceNoul)),
		O.Map(F.Constant1[components.DecisionsResponseAnswers](label)),
	)
}

func extractAnswersLabels(answers answersMap) []string {
	return F.Pipe2(
		jevAnswerExtractors,
		A.Flap[O.Option[string]](answers),
		arrutils.Compact[string],
	)
}

func deduplicateLabels(labels []string) []string {
	return F.Pipe3(
		labels,
		A.Map(strings.TrimSpace),
		A.Filter(S.IsNonEmpty),
		A.StrictUniq[string],
	)
}

func formatUSStateDescription(code, stateName string) string {
	return strutils.JoinStrings(
		"US shipping address in ",
		stateName,
		" ",
		strutils.Parens(code),
		".",
	)
}

func intlLocationLabels() map[string]string {
	return map[string]string{
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
}

func locationCriteriaMap() map[string]string {
	return F.Pipe2(
		usStateLabels(),
		R.MapWithIndex(formatUSStateDescription),
		F.Bind2nd(
			R.UnionLastMonoid[string, string]().Concat,
			intlLocationLabels(),
		),
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
