package github

import (
	"context"
	"errors"
	"io"
	"testing"

	F "github.com/IBM/fp-go/v2/function"
	L "github.com/IBM/fp-go/v2/optics/lens"
	"github.com/OpenRouterTeam/go-sdk/models/components"
	"github.com/OpenRouterTeam/go-sdk/models/operations"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const testStrainName = "AX2"

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
		ordAttrItemsLens.Set([]string{"DBS0237699", "DBP0000105"}),
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
			{"K. aerogenes GFP", "bacterial stock", "21(17)", "2", "yellow"},
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

	got := gh.resolveLabels(new(all), sampleOrder())
	require.ElementsMatch(
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
			[][]string{{"DBS0237699", testStrainName, testStrainName, testStrainName, ""}},
		),
		strainDataInvsLens.Set(
			[][]string{{testStrainName, "bacterial stock", "21(17)", "2", "yellow"}},
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

	got := gh.resolveLabels(new(all), sampleOrder())

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
	require.Equal(t, []string{labelPlasmidOrder}, ghLow.resolveLabels(new(all), sampleOrder()))

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
	require.Equal(t, []string{labelPlasmidOrder}, ghErr.resolveLabels(new(all), sampleOrder()))
}
