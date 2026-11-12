package github

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	J "github.com/IBM/fp-go/v2/json"
	L "github.com/IBM/fp-go/v2/optics/lens"
	"github.com/OpenRouterTeam/go-sdk/models/components"
	"github.com/OpenRouterTeam/go-sdk/models/operations"
	"github.com/dictyBase/event-messenger/internal/datasource"
	"github.com/dictyBase/event-messenger/internal/fake"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictyBase/go-genproto/dictybaseapis/annotation"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	ghapi "github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

const (
	testStrainName     = "AX2"
	testStrainID       = "DBS0237699"
	testPlasmidID      = "DBP0000105"
	testBacterialStock = "bacterial stock"
)

type fakeDecisionsClient struct {
	resp  *components.DecisionsResponse
	err   error
	calls int
	last  components.DecisionsRequest
}

var (
	fakeRespLens = L.MakeLens(
		func(f fakeDecisionsClient) *components.DecisionsResponse { return f.resp },
		func(f fakeDecisionsClient, v *components.DecisionsResponse) fakeDecisionsClient {
			f.resp = v
			return f
		},
	)

	fakeErrLens = L.MakeLens(
		func(f fakeDecisionsClient) error { return f.err },
		func(f fakeDecisionsClient, v error) fakeDecisionsClient {
			f.err = v
			return f
		},
	)

	fakeCallsLens = L.MakeLens(
		func(f fakeDecisionsClient) int { return f.calls },
		func(f fakeDecisionsClient, v int) fakeDecisionsClient {
			f.calls = v
			return f
		},
	)

	fakeLastReqLens = L.MakeLens(
		func(f fakeDecisionsClient) components.DecisionsRequest { return f.last },
		func(f fakeDecisionsClient, v components.DecisionsRequest) fakeDecisionsClient {
			f.last = v
			return f
		},
	)

	respAnswersLens = L.MakeLens(
		func(
			r components.DecisionsResponse,
		) map[string]components.DecisionsResponseAnswers {
			return r.Answers
		},
		func(
			r components.DecisionsResponse,
			v map[string]components.DecisionsResponseAnswers,
		) components.DecisionsResponse {
			r.Answers = v
			return r
		},
	)

	choiceAnsChoiceLens = L.MakeLens(
		func(a components.DecisionsChoiceAnswer) string { return a.Choice },
		func(
			a components.DecisionsChoiceAnswer,
			v string,
		) components.DecisionsChoiceAnswer {
			a.Choice = v
			return a
		},
	)

	choiceAnsConfLens = L.MakeLens(
		func(a components.DecisionsChoiceAnswer) *float64 { return a.Confidence },
		func(
			a components.DecisionsChoiceAnswer,
			v *float64,
		) components.DecisionsChoiceAnswer {
			a.Confidence = v
			return a
		},
	)

	choiceAnsProbsLens = L.MakeLens(
		func(a components.DecisionsChoiceAnswer) map[string]float64 {
			return a.Probabilities
		},
		func(
			a components.DecisionsChoiceAnswer,
			v map[string]float64,
		) components.DecisionsChoiceAnswer {
			a.Probabilities = v
			return a
		},
	)

	choiceAnsTypeLens = L.MakeLens(
		func(
			a components.DecisionsChoiceAnswer,
		) components.DecisionsChoiceAnswerType {
			return a.Type
		},
		func(
			a components.DecisionsChoiceAnswer,
			v components.DecisionsChoiceAnswerType,
		) components.DecisionsChoiceAnswer {
			a.Type = v
			return a
		},
	)

	noulAnsValLens = L.MakeLens(
		func(a components.DecisionsNoulAnswer) float64 { return a.Noul },
		func(
			a components.DecisionsNoulAnswer,
			v float64,
		) components.DecisionsNoulAnswer {
			a.Noul = v
			return a
		},
	)

	noulAnsTypeLens = L.MakeLens(
		func(
			a components.DecisionsNoulAnswer,
		) components.DecisionsNoulAnswerType {
			return a.Type
		},
		func(
			a components.DecisionsNoulAnswer,
			v components.DecisionsNoulAnswerType,
		) components.DecisionsNoulAnswer {
			a.Type = v
			return a
		},
	)

	userFirstLens = L.MakeLens(
		func(u *order.UserInfo) string { return u.GetFirstName() },
		func(u *order.UserInfo, v string) *order.UserInfo {
			u.FirstName = v
			return u
		},
	)

	userLastLens = L.MakeLens(
		func(u *order.UserInfo) string { return u.GetLastName() },
		func(u *order.UserInfo, v string) *order.UserInfo {
			u.LastName = v
			return u
		},
	)

	userOrgLens = L.MakeLens(
		func(u *order.UserInfo) string { return u.GetOrganization() },
		func(u *order.UserInfo, v string) *order.UserInfo {
			u.Organization = v
			return u
		},
	)

	userAddrLens = L.MakeLens(
		func(u *order.UserInfo) string { return u.GetFirstAddress() },
		func(u *order.UserInfo, v string) *order.UserInfo {
			u.FirstAddress = v
			return u
		},
	)

	userCityLens = L.MakeLens(
		func(u *order.UserInfo) string { return u.GetCity() },
		func(u *order.UserInfo, v string) *order.UserInfo {
			u.City = v
			return u
		},
	)

	userStateLens = L.MakeLens(
		func(u *order.UserInfo) string { return u.GetState() },
		func(u *order.UserInfo, v string) *order.UserInfo {
			u.State = v
			return u
		},
	)

	userZipLens = L.MakeLens(
		func(u *order.UserInfo) string { return u.GetZipcode() },
		func(u *order.UserInfo, v string) *order.UserInfo {
			u.Zipcode = v
			return u
		},
	)

	userCountryLens = L.MakeLens(
		func(u *order.UserInfo) string { return u.GetCountry() },
		func(u *order.UserInfo, v string) *order.UserInfo {
			u.Country = v
			return u
		},
	)

	ordAttrPurchaserLens = L.MakeLens(
		func(a *order.OrderAttributes) string { return a.GetPurchaser() },
		func(a *order.OrderAttributes, v string) *order.OrderAttributes {
			a.Purchaser = v
			return a
		},
	)

	ordAttrConsumerLens = L.MakeLens(
		func(a *order.OrderAttributes) string { return a.GetConsumer() },
		func(a *order.OrderAttributes, v string) *order.OrderAttributes {
			a.Consumer = v
			return a
		},
	)

	ordAttrPayerLens = L.MakeLens(
		func(a *order.OrderAttributes) string { return a.GetPayer() },
		func(a *order.OrderAttributes, v string) *order.OrderAttributes {
			a.Payer = v
			return a
		},
	)

	ordAttrPaymentLens = L.MakeLens(
		func(a *order.OrderAttributes) string { return a.GetPayment() },
		func(a *order.OrderAttributes, v string) *order.OrderAttributes {
			a.Payment = v
			return a
		},
	)

	ordAttrPOLens = L.MakeLens(
		func(a *order.OrderAttributes) string { return a.GetPurchaseOrderNum() },
		func(a *order.OrderAttributes, v string) *order.OrderAttributes {
			a.PurchaseOrderNum = v
			return a
		},
	)

	ordAttrCourierLens = L.MakeLens(
		func(a *order.OrderAttributes) string { return a.GetCourier() },
		func(a *order.OrderAttributes, v string) *order.OrderAttributes {
			a.Courier = v
			return a
		},
	)

	ordAttrCourierAcctLens = L.MakeLens(
		func(a *order.OrderAttributes) string { return a.GetCourierAccount() },
		func(a *order.OrderAttributes, v string) *order.OrderAttributes {
			a.CourierAccount = v
			return a
		},
	)

	ordAttrCommentsLens = L.MakeLens(
		func(a *order.OrderAttributes) string { return a.GetComments() },
		func(a *order.OrderAttributes, v string) *order.OrderAttributes {
			a.Comments = v
			return a
		},
	)

	ordAttrItemsLens = L.MakeLens(
		func(a *order.OrderAttributes) []string { return a.GetItems() },
		func(a *order.OrderAttributes, v []string) *order.OrderAttributes {
			a.Items = v
			return a
		},
	)

	ordAttrConsumerInfoLens = L.MakeLens(
		func(a *order.OrderAttributes) *order.UserInfo { return a.GetConsumerInfo() },
		func(a *order.OrderAttributes, v *order.UserInfo) *order.OrderAttributes {
			a.ConsumerInfo = v
			return a
		},
	)

	ordDataIDLens = L.MakeLens(
		func(d *order.Order_Data) string { return d.GetId() },
		func(d *order.Order_Data, v string) *order.Order_Data {
			d.Id = v
			return d
		},
	)

	ordDataAttrLens = L.MakeLens(
		func(d *order.Order_Data) *order.OrderAttributes { return d.GetAttributes() },
		func(d *order.Order_Data, v *order.OrderAttributes) *order.Order_Data {
			d.Attributes = v
			return d
		},
	)

	ordRootDataLens = L.MakeLens(
		func(o *order.Order) *order.Order_Data { return o.GetData() },
		func(o *order.Order, v *order.Order_Data) *order.Order {
			o.Data = v
			return o
		},
	)
)

func (f *fakeDecisionsClient) Create(
	_ context.Context,
	req components.DecisionsRequest,
	_ ...operations.Option,
) (*components.DecisionsResponse, error) {
	updated := F.Pipe2(
		*f,
		fakeCallsLens.Set(fakeCallsLens.Get(*f)+1),
		fakeLastReqLens.Set(req),
	)
	*f = updated

	return fakeRespLens.Get(*f), fakeErrLens.Get(*f)
}

func makeChoiceAns(choice string, conf float64) components.DecisionsResponseAnswers {
	ans := F.Pipe3(
		components.DecisionsChoiceAnswer{},
		choiceAnsChoiceLens.Set(choice),
		choiceAnsConfLens.Set(new(conf)),
		choiceAnsTypeLens.Set(components.DecisionsChoiceAnswerTypeChoice),
	)

	return components.CreateDecisionsResponseAnswersChoice(ans)
}

func makeNoulAns(prob float64) components.DecisionsResponseAnswers {
	ans := F.Pipe2(
		components.DecisionsNoulAnswer{},
		noulAnsValLens.Set(prob),
		noulAnsTypeLens.Set(components.DecisionsNoulAnswerTypeNoul),
	)

	return components.CreateDecisionsResponseAnswersNoul(ans)
}

func testLogger() *logrus.Entry {
	lg := logrus.New()
	lg.SetOutput(io.Discard)

	return logrus.NewEntry(lg)
}

func sampleOrder() *order.Order {
	consumer := F.Pipe8(
		new(order.UserInfo),
		userFirstLens.Set("Michael"),
		userLastLens.Set("Edgar"),
		userOrgLens.Set("Milton Academy"),
		userAddrLens.Set("170 Centre Street"),
		userCityLens.Set("Milton"),
		userStateLens.Set("MA"),
		userZipLens.Set("02186"),
		userCountryLens.Set("USA"),
	)
	attr := F.Pipe10(
		new(order.OrderAttributes),
		ordAttrPurchaserLens.Set("Michael Edgar"),
		ordAttrConsumerLens.Set("michael_edgar@milton.edu"),
		ordAttrPayerLens.Set("michael_edgar@milton.edu"),
		ordAttrPaymentLens.Set("Credit card"),
		ordAttrPOLens.Set("N/A"),
		ordAttrCourierLens.Set("FedEx"),
		ordAttrCourierAcctLens.Set("2024-5656-1"),
		ordAttrCommentsLens.Set("please include bacterial food source"),
		ordAttrItemsLens.Set([]string{testStrainID, testPlasmidID}),
		ordAttrConsumerInfoLens.Set(consumer),
	)
	data := F.Pipe2(
		new(order.Order_Data),
		ordDataIDLens.Set("18776"),
		ordDataAttrLens.Set(attr),
	)

	return F.Pipe1(
		new(order.Order),
		ordRootDataLens.Set(data),
	)
}

func makeCreateState(all *allData, ord *order.Order) issueCreateState {
	return F.Pipe2(
		issueCreateState{},
		createStateDataLens.Set(all),
		createStateOrderLens.Set(ord),
	)
}

func mustResolveLabels(
	t *testing.T,
	gh *githubIssue,
	all *allData,
	ord *order.Order,
) []string {
	t.Helper()

	res := ioeutils.ToEither(gh.resolveLabels(makeCreateState(all, ord)))
	require.True(t, E.IsRight(res))

	return F.Pipe1(res, E.GetOrElse(F.Constant1[error]([]string(nil))))
}

func TestDeterministicLabels(t *testing.T) {
	t.Parallel()

	sData := F.Pipe3(
		strainData{},
		strainDataItemsLens.Set([]*stock.Strain{{}}),
		strainDataInfoLens.Set([][]string{
			{"DBS0350001", "GWDI_28_G_7", "gwdi", "GWDI", "blasticidin resistant"},
		}),
		strainDataInvsLens.Set([][]string{
			{"GWDI_28_G_7", "axenic cells", "1-2(3)", "3", "pink"},
			{"K. aerogenes GFP", testBacterialStock, "21(17)", "2", "yellow"},
		}),
	)
	pData := F.Pipe2(
		plasmidData{},
		plasmidDataItemsLens.Set([]*stock.Plasmid{{}}),
		plasmidDataInvsLens.Set([][]string{
			{"DBP0001001", "pDGB_A1_0051", "DH5a", "26(9)", "blue"},
		}),
	)
	all := F.Pipe2(
		allData{},
		allStrainDataLens.Set(new(sData)),
		allPlasmidDataLens.Set(new(pData)),
	)
	gh := F.Pipe1(githubIssue{}, ghLoggerLens.Set(testLogger()))

	got := mustResolveLabels(t, new(gh), new(all), sampleOrder())
	require.Equal(
		t,
		[]string{
			labelStrainOrder,
			labelPlasmidOrder,
			labelAxenicMedium,
			labelGWDI,
			labelBacterialFood,
			labelGoldenBraid,
		},
		got,
	)
}

func TestJevLabelsEnrichmentAndDeduplication(t *testing.T) {
	t.Parallel()

	resp := F.Pipe1(
		components.DecisionsResponse{},
		respAnswersLens.Set(map[string]components.DecisionsResponseAnswers{
			questionInstitutionType: makeChoiceAns(labelResearchInst, 0.95),
			questionIsMSI:           makeNoulAns(0.91),
			questionPaymentLabel:    makeChoiceAns(labelCreditCard, 0.95),
			questionLocationLabel:   makeChoiceAns("MA", 0.95),
			questionBacterialFood:   makeNoulAns(0.88),
			questionGoldenBraid:     makeNoulAns(0.10),
			questionGWDI:            makeNoulAns(0.05),
			questionAxenicMedium:    makeNoulAns(0.20),
		}),
	)
	fakeVal := F.Pipe1(fakeDecisionsClient{}, fakeRespLens.Set(new(resp)))
	fakeClient := new(fakeVal)

	sData := F.Pipe3(
		strainData{},
		strainDataItemsLens.Set([]*stock.Strain{{}}),
		strainDataInfoLens.Set(
			[][]string{
				{testStrainID, testStrainName, testStrainName, testStrainName, ""},
			},
		),
		strainDataInvsLens.Set(
			[][]string{{testStrainName, testBacterialStock, "21(17)", "2", "yellow"}},
		),
	)
	all := F.Pipe2(
		allData{},
		allStrainDataLens.Set(new(sData)),
		allPlasmidDataLens.Set(new(plasmidData{})),
	)
	gh := F.Pipe3(
		githubIssue{},
		ghLoggerLens.Set(testLogger()),
		ghDecisionsLens.Set(DecisionsClient(fakeClient)),
		ghJevModelLens.Set(defaultJevModel),
	)

	got := mustResolveLabels(t, new(gh), new(all), sampleOrder())

	require.Equal(t, 1, fakeCallsLens.Get(*fakeClient))
	require.Equal(t, defaultJevModel, decReqModelLens.Get(fakeLastReqLens.Get(*fakeClient)))
	require.Equal(
		t,
		[]string{
			labelStrainOrder,
			labelBacterialFood,
			labelResearchInst,
			labelCreditCard,
			"MA",
			labelMSI,
		},
		got,
	)
}

func TestJevLabelsLowConfidenceAndErrorFallback(t *testing.T) {
	t.Parallel()

	lowResp := F.Pipe1(
		components.DecisionsResponse{},
		respAnswersLens.Set(map[string]components.DecisionsResponseAnswers{
			questionInstitutionType: makeChoiceAns(labelHighSchool, 0.40),
			questionLocationLabel:   makeChoiceAns(choiceNone, 0.99),
		}),
	)
	lowClientVal := F.Pipe1(fakeDecisionsClient{}, fakeRespLens.Set(new(lowResp)))

	pData := F.Pipe1(
		plasmidData{},
		plasmidDataItemsLens.Set([]*stock.Plasmid{{}}),
	)
	all := F.Pipe2(
		allData{},
		allStrainDataLens.Set(new(strainData{})),
		allPlasmidDataLens.Set(new(pData)),
	)

	ghLow := F.Pipe3(
		githubIssue{},
		ghLoggerLens.Set(testLogger()),
		ghDecisionsLens.Set(DecisionsClient(new(lowClientVal))),
		ghJevModelLens.Set(defaultJevModel),
	)
	require.Equal(
		t,
		[]string{labelPlasmidOrder},
		mustResolveLabels(t, new(ghLow), new(all), sampleOrder()),
	)

	errClientVal := F.Pipe1(
		fakeDecisionsClient{},
		fakeErrLens.Set(errors.New("openrouter timeout")),
	)
	ghErr := F.Pipe3(
		githubIssue{},
		ghLoggerLens.Set(testLogger()),
		ghDecisionsLens.Set(DecisionsClient(new(errClientVal))),
		ghJevModelLens.Set(defaultJevModel),
	)
	require.Equal(
		t,
		[]string{labelPlasmidOrder},
		mustResolveLabels(t, new(ghErr), new(all), sampleOrder()),
	)
}

func makeProbChoiceAns(
	choice string,
	probs map[string]float64,
) components.DecisionsResponseAnswers {
	ans := F.Pipe3(
		components.DecisionsChoiceAnswer{},
		choiceAnsChoiceLens.Set(choice),
		choiceAnsProbsLens.Set(probs),
		choiceAnsTypeLens.Set(components.DecisionsChoiceAnswerTypeChoice),
	)

	return components.CreateDecisionsResponseAnswersChoice(ans)
}

func TestDeduplicateAndEmptySliceBehavior(t *testing.T) {
	t.Parallel()

	require.Nil(t, deduplicateLabels(nil))
	require.Nil(t, deduplicateLabels([]string{}))
	require.Nil(t, deduplicateLabels([]string{"", "   ", "\t"}))
	require.Equal(
		t,
		[]string{"MA", labelPurchaseOrder, labelCommercial},
		deduplicateLabels([]string{
			" MA ",
			labelPurchaseOrder,
			"MA",
			" ",
			labelCommercial,
			labelPurchaseOrder,
		}),
	)
	require.Nil(t, deterministicLabels(nil))
	require.Nil(t, deterministicLabels(&allData{}))
}

func TestJevConfidenceBoundaryAndProbabilityFallback(t *testing.T) {
	t.Parallel()

	resp := F.Pipe1(
		components.DecisionsResponse{},
		respAnswersLens.Set(map[string]components.DecisionsResponseAnswers{
			questionInstitutionType: makeChoiceAns(" "+labelHighSchool+" ", 0.75),
			questionPaymentLabel:    makeChoiceAns(labelPurchaseOrder, 0.749),
			questionLocationLabel: makeProbChoiceAns(
				" CA ",
				map[string]float64{"CA": 0.75, "TX": 0.25},
			),
			questionIsMSI:        makeNoulAns(0.75),
			questionAxenicMedium: makeNoulAns(0.749),
		}),
	)
	fakeVal := F.Pipe1(fakeDecisionsClient{}, fakeRespLens.Set(new(resp)))
	gh := F.Pipe3(
		githubIssue{},
		ghLoggerLens.Set(testLogger()),
		ghDecisionsLens.Set(DecisionsClient(new(fakeVal))),
		ghJevModelLens.Set("   "),
	)

	got := mustResolveLabels(t, new(gh), &allData{}, sampleOrder())
	require.Equal(t, []string{labelHighSchool, "CA", labelMSI}, got)

	lowProbResp := F.Pipe1(
		components.DecisionsResponse{},
		respAnswersLens.Set(map[string]components.DecisionsResponseAnswers{
			questionLocationLabel: makeProbChoiceAns(
				"TX",
				map[string]float64{"TX": 0.749},
			),
		}),
	)
	lowProbVal := F.Pipe1(
		fakeDecisionsClient{},
		fakeRespLens.Set(new(lowProbResp)),
	)
	ghLowProb := F.Pipe2(
		githubIssue{},
		ghLoggerLens.Set(testLogger()),
		ghDecisionsLens.Set(DecisionsClient(new(lowProbVal))),
	)
	require.Nil(t, mustResolveLabels(t, new(ghLowProb), &allData{}, sampleOrder()))
}

func TestJevNilResponseAndLoggerWarningBehavior(t *testing.T) {
	t.Parallel()

	nilRespClient := new(fakeDecisionsClient{})
	pData := F.Pipe1(
		plasmidData{},
		plasmidDataItemsLens.Set([]*stock.Plasmid{{}}),
	)
	all := F.Pipe2(
		allData{},
		allStrainDataLens.Set(new(strainData{})),
		allPlasmidDataLens.Set(new(pData)),
	)
	ghNilResp := F.Pipe2(
		githubIssue{},
		ghLoggerLens.Set(testLogger()),
		ghDecisionsLens.Set(DecisionsClient(nilRespClient)),
	)
	require.Equal(
		t,
		[]string{labelPlasmidOrder},
		mustResolveLabels(t, new(ghNilResp), new(all), sampleOrder()),
	)

	errClientNilLog := F.Pipe1(
		fakeDecisionsClient{},
		fakeErrLens.Set(errors.New("openrouter timeout")),
	)
	ghNilLog := F.Pipe1(
		githubIssue{},
		ghDecisionsLens.Set(DecisionsClient(new(errClientNilLog))),
	)
	require.Equal(
		t,
		[]string{labelPlasmidOrder},
		mustResolveLabels(t, new(ghNilLog), new(all), sampleOrder()),
	)

	var logBuf bytes.Buffer

	lg := logrus.New()
	lg.SetOutput(&logBuf)

	errClientBufLog := F.Pipe1(
		fakeDecisionsClient{},
		fakeErrLens.Set(errors.New("openrouter timeout")),
	)
	ghBufLog := F.Pipe2(
		githubIssue{},
		ghLoggerLens.Set(logrus.NewEntry(lg)),
		ghDecisionsLens.Set(DecisionsClient(new(errClientBufLog))),
	)
	require.Equal(
		t,
		[]string{labelPlasmidOrder},
		mustResolveLabels(t, new(ghBufLog), new(all), sampleOrder()),
	)
	require.Contains(
		t,
		logBuf.String(),
		"openrouter jev label classification failed: openrouter timeout",
	)
}

func TestBuildDecisionsStatePayloadAndQuestions(t *testing.T) {
	t.Parallel()

	sData := F.Pipe2(
		strainData{},
		strainDataInfoLens.Set([][]string{{testStrainID, testStrainName}}),
		strainDataInvsLens.Set([][]string{{testStrainName, testBacterialStock}}),
	)
	pData := F.Pipe1(
		plasmidData{},
		plasmidDataInvsLens.Set([][]string{{testPlasmidID, "pDGB_A1"}}),
	)
	all := F.Pipe2(
		allData{},
		allStrainDataLens.Set(new(sData)),
		allPlasmidDataLens.Set(new(pData)),
	)

	stateMap := buildDecisionsState(new(all), sampleOrder())
	require.Equal(t, "18776", stateMap["order_id"])
	require.Equal(t, "Michael Edgar", stateMap["purchaser"])
	require.Equal(t, "michael_edgar@milton.edu", stateMap["consumer_email"])
	require.Equal(t, "michael_edgar@milton.edu", stateMap["payer_email"])
	require.Equal(t, "Credit card", stateMap["payment_method"])
	require.Equal(t, "N/A", stateMap["purchase_order_num"])
	require.Equal(t, "FedEx", stateMap["courier"])
	require.Equal(
		t,
		"please include bacterial food source",
		stateMap["comments"],
	)
	require.Equal(t, []string{testStrainID, testPlasmidID}, stateMap["items"])
	shippingJSON := J.Marshal(stateMap["shipping_address"])
	require.True(t, E.IsRight(shippingJSON))
	require.JSONEq(
		t,
		`{
			"first_name": "Michael",
			"last_name": "Edgar",
			"organization": "Milton Academy",
			"first_address": "170 Centre Street",
			"second_address": "",
			"city": "Milton",
			"state": "MA",
			"zipcode": "02186",
			"country": "USA"
		}`,
		string(F.Pipe1(shippingJSON, E.GetOrElse(F.Constant1[error]([]byte(nil))))),
	)

	billingJSON := J.Marshal(stateMap["billing_address"])
	require.True(t, E.IsRight(billingJSON))
	require.JSONEq(
		t,
		`{
			"organization": "",
			"city": "",
			"state": "",
			"country": ""
		}`,
		string(F.Pipe1(billingJSON, E.GetOrElse(F.Constant1[error]([]byte(nil))))),
	)
	require.Equal(
		t,
		[][]string{{testStrainID, testStrainName}},
		stateMap["strain_info"],
	)
	require.Equal(
		t,
		[][]string{{testStrainName, testBacterialStock}},
		stateMap["strain_inventory"],
	)
	require.Equal(
		t,
		[][]string{{testPlasmidID, "pDGB_A1"}},
		stateMap["plasmid_inventory"],
	)

	questions := buildDecisionsQuestions()
	require.Len(t, questions, 8)
	require.Contains(t, questions, questionInstitutionType)
	require.Contains(t, questions, questionIsMSI)
	require.Contains(t, questions, questionPaymentLabel)
	require.Contains(t, questions, questionLocationLabel)
	require.Contains(t, questions, questionBacterialFood)
	require.Contains(t, questions, questionGoldenBraid)
	require.Contains(t, questions, questionGWDI)
	require.Contains(t, questions, questionAxenicMedium)
}

func TestPreparePostParamsIsPureWithoutJevCall(t *testing.T) {
	t.Parallel()

	resp := F.Pipe1(
		components.DecisionsResponse{},
		respAnswersLens.Set(map[string]components.DecisionsResponseAnswers{
			questionInstitutionType: makeChoiceAns(labelResearchInst, 0.95),
		}),
	)
	fakeVal := F.Pipe1(fakeDecisionsClient{}, fakeRespLens.Set(new(resp)))
	fakeClient := new(fakeVal)

	gh := F.Pipe2(
		githubIssue{},
		ghLoggerLens.Set(testLogger()),
		ghDecisionsLens.Set(DecisionsClient(fakeClient)),
	)
	state := F.Pipe3(
		issueCreateState{},
		createStateOrderLens.Set(sampleOrder()),
		createStateDataLens.Set(&allData{}),
		createStateBodyLens.Set(bytes.NewBufferString("rendered body")),
	)

	params := gh.preparePostParams(state)
	require.Equal(t, "Order ID:18776 Michael Edgar", postTitleLens.Get(*params))
	require.Equal(t, "rendered body", postBodyLens.Get(*params))
	require.Equal(t, 0, fakeCallsLens.Get(*fakeClient))
}

func TestResolveLabelsIsLazyUntilForced(t *testing.T) {
	t.Parallel()

	resp := F.Pipe1(
		components.DecisionsResponse{},
		respAnswersLens.Set(map[string]components.DecisionsResponseAnswers{
			questionInstitutionType: makeChoiceAns(labelResearchInst, 0.95),
		}),
	)
	fakeVal := F.Pipe1(fakeDecisionsClient{}, fakeRespLens.Set(new(resp)))
	fakeClient := new(fakeVal)

	ghVal := F.Pipe2(
		githubIssue{},
		ghLoggerLens.Set(testLogger()),
		ghDecisionsLens.Set(DecisionsClient(fakeClient)),
	)
	gh := new(ghVal)

	effect := gh.resolveLabels(makeCreateState(&allData{}, sampleOrder()))

	require.Equal(t, 0, fakeCallsLens.Get(*fakeClient))

	forced := ioeutils.ToEither(effect)
	require.True(t, E.IsRight(forced))
	require.Equal(t, 1, fakeCallsLens.Get(*fakeClient))
	require.Equal(
		t,
		[]string{labelResearchInst},
		F.Pipe1(forced, E.GetOrElse(F.Constant1[error]([]string(nil)))),
	)
}

type stubStockClient struct {
	stock.StockServiceClient
}

func (stubStockClient) GetStrain(
	_ context.Context,
	req *stock.StockId,
	_ ...grpc.CallOption,
) (*stock.Strain, error) {
	st := fake.Strain()
	st.Data.Id = req.GetId()

	return st, nil
}

func (stubStockClient) GetPlasmid(
	_ context.Context,
	req *stock.StockId,
	_ ...grpc.CallOption,
) (*stock.Plasmid, error) {
	pls := fake.Plasmid()
	pls.Data.Id = req.GetId()

	return pls, nil
}

type stubAnnoClient struct {
	annotation.TaggedAnnotationServiceClient
}

func (stubAnnoClient) ListAnnotationGroups(
	_ context.Context,
	req *annotation.ListGroupParameters,
	_ ...grpc.CallOption,
) (*annotation.TaggedAnnotationGroupCollection, error) {
	if strings.Contains(req.GetFilter(), "plasmid_inventory") {
		return fake.PlasmidInvAnno(), nil
	}

	return fake.StrainInvAnno(), nil
}

func (stubAnnoClient) ListAnnotations(
	_ context.Context,
	_ *annotation.ListParameters,
	_ ...grpc.CallOption,
) (*annotation.TaggedAnnotationCollection, error) {
	return &annotation.TaggedAnnotationCollection{}, nil
}

func (stubAnnoClient) GetEntryAnnotation(
	_ context.Context,
	_ *annotation.EntryAnnotationRequest,
	_ ...grpc.CallOption,
) (*annotation.TaggedAnnotation, error) {
	return fake.SysNameAnno(), nil
}

var (
	stockClientLens = L.MakeLensRef(
		func(s *datasource.Stock) stock.StockServiceClient { return s.Client },
		func(
			s *datasource.Stock,
			v stock.StockServiceClient,
		) *datasource.Stock {
			s.Client = v
			return s
		},
	)

	annoClientLens = L.MakeLensRef(
		func(a *datasource.Annotation) annotation.TaggedAnnotationServiceClient {
			return a.Client
		},
		func(
			a *datasource.Annotation,
			v annotation.TaggedAnnotationServiceClient,
		) *datasource.Annotation {
			a.Client = v
			return a
		},
	)

	testSourcesStockLens = L.MakeLensRef(
		func(s *datasource.Sources) *datasource.Stock { return s.StockSource },
		func(s *datasource.Sources, v *datasource.Stock) *datasource.Sources {
			s.StockSource = v
			return s
		},
	)

	testSourcesAnnoLens = L.MakeLensRef(
		func(s *datasource.Sources) *datasource.Annotation { return s.AnnoSource },
		func(
			s *datasource.Sources,
			v *datasource.Annotation,
		) *datasource.Sources {
			s.AnnoSource = v
			return s
		},
	)

	testParamsSourcesLens = L.MakeLens(
		func(p *IssueParams) *datasource.Sources { return p.Sources },
		func(p *IssueParams, v *datasource.Sources) *IssueParams {
			p.Sources = v
			return p
		},
	)
)

func testSources() *datasource.Sources {
	stk := F.Pipe1(&datasource.Stock{}, stockClientLens.Set(stubStockClient{}))
	anno := F.Pipe1(
		&datasource.Annotation{},
		annoClientLens.Set(stubAnnoClient{}),
	)

	return F.Pipe2(
		&datasource.Sources{},
		testSourcesStockLens.Set(stk),
		testSourcesAnnoLens.Set(anno),
	)
}

func newTestIssueTracker(
	t *testing.T,
	srvURL string,
	decisions DecisionsClient,
) *githubIssue {
	t.Helper()

	params := F.Pipe9(
		new(IssueParams),
		issueParamsTokenLens.Set("gh-test-token"),
		issueParamsOwnerLens.Set("dictyBase"),
		issueParamsRepoLens.Set("Stock-Center-Orders"),
		issueParamsOpenRouterLens.Set("ignored-when-decisions-injected"),
		issueParamsDecisionsLens.Set(decisions),
		issueParamsLoggerLens.Set(testLogger()),
		issueParamsStrainPriceLens.Set(30),
		issueParamsPlasmidPriceLens.Set(15),
		testParamsSourcesLens.Set(testSources()),
	)
	tracker := NewIssueCreator(params)
	gh, ok := tracker.(*githubIssue)
	require.True(t, ok)

	baseURL, err := url.Parse(srvURL + "/")
	require.NoError(t, err)

	ghClientLens.Get(*gh).BaseURL = baseURL

	return gh
}

func TestCreateIssueEnrichesLabelsViaJev(t *testing.T) {
	t.Parallel()

	var (
		postCalls int
		gotRaw    []byte
		readErr   error
	)

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			postCalls++
			gotRaw, readErr = io.ReadAll(r.Body)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write(
				[]byte(
					`{"html_url":"https://github.com/dictyBase/Stock-Center-Orders/issues/41"}`,
				),
			)
		}),
	)
	defer srv.Close()

	resp := F.Pipe1(
		components.DecisionsResponse{},
		respAnswersLens.Set(map[string]components.DecisionsResponseAnswers{
			questionInstitutionType: makeChoiceAns(labelHighSchool, 0.96),
			questionPaymentLabel:    makeChoiceAns(labelCreditCard, 0.94),
			questionLocationLabel:   makeChoiceAns("MA", 0.98),
		}),
	)
	fakeVal := F.Pipe1(fakeDecisionsClient{}, fakeRespLens.Set(new(resp)))
	fakeClient := new(fakeVal)

	gh := newTestIssueTracker(t, srv.URL, fakeClient)

	err := gh.CreateIssue(sampleOrder())
	require.NoError(t, err)
	require.Equal(t, 1, fakeCallsLens.Get(*fakeClient))
	require.Equal(t, 1, postCalls)
	require.NoError(t, readErr)

	parsed := J.Unmarshal[ghapi.IssueRequest](gotRaw)
	require.True(t, E.IsRight(parsed))
	gotReq := F.Pipe1(
		parsed,
		E.GetOrElse(F.Constant1[error](ghapi.IssueRequest{})),
	)

	require.Equal(t, "Order ID:18776 Michael Edgar", gotReq.GetTitle())
	require.Contains(t, gotReq.GetBody(), "Milton Academy")
	require.Equal(
		t,
		[]string{
			labelStrainOrder,
			labelPlasmidOrder,
			labelAxenicMedium,
			labelHighSchool,
			labelCreditCard,
			"MA",
		},
		*gotReq.Labels,
	)
}

func TestCreateIssueRecoversJevFailureAndPostsOnce(t *testing.T) {
	t.Parallel()

	var (
		postCalls int
		gotMethod string
		gotPath   string
		gotRaw    []byte
		readErr   error
	)

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			postCalls++
			gotMethod = r.Method
			gotPath = r.URL.Path
			gotRaw, readErr = io.ReadAll(r.Body)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write(
				[]byte(
					`{"html_url":"https://github.com/dictyBase/Stock-Center-Orders/issues/42"}`,
				),
			)
		}),
	)
	defer srv.Close()

	errVal := F.Pipe1(
		fakeDecisionsClient{},
		fakeErrLens.Set(errors.New("openrouter 503")),
	)
	fakeClient := new(errVal)

	gh := newTestIssueTracker(t, srv.URL, fakeClient)

	err := gh.CreateIssue(sampleOrder())
	require.NoError(t, err)
	require.Equal(t, 1, fakeCallsLens.Get(*fakeClient))
	require.Equal(t, 1, postCalls)
	require.Equal(t, http.MethodPost, gotMethod)
	require.Equal(t, "/repos/dictyBase/Stock-Center-Orders/issues", gotPath)
	require.NoError(t, readErr)

	parsed := J.Unmarshal[ghapi.IssueRequest](gotRaw)
	require.True(t, E.IsRight(parsed))
	gotReq := F.Pipe1(
		parsed,
		E.GetOrElse(F.Constant1[error](ghapi.IssueRequest{})),
	)

	require.Equal(t, "Order ID:18776 Michael Edgar", gotReq.GetTitle())
	require.Contains(t, gotReq.GetBody(), "Milton Academy")
	require.Contains(t, gotReq.GetBody(), "p123456")
	require.Equal(
		t,
		[]string{labelStrainOrder, labelPlasmidOrder, labelAxenicMedium},
		*gotReq.Labels,
	)
}
