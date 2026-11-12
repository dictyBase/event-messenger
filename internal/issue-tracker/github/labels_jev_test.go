package github

import (
	"context"
	"errors"
	"io"
	"testing"

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

func (f *fakeDecisionsClient) Create(
	_ context.Context,
	req components.DecisionsRequest,
	_ ...operations.Option,
) (*components.DecisionsResponse, error) {
	f.calls++
	f.last = req

	return f.resp, f.err
}

func testLogger() *logrus.Entry {
	lg := logrus.New()
	lg.SetOutput(io.Discard)

	return logrus.NewEntry(lg)
}

func sampleOrder() *order.Order {
	return &order.Order{
		Data: &order.Order_Data{
			Id: "18776",
			Attributes: &order.OrderAttributes{
				Purchaser:        "Michael Edgar",
				Consumer:         "michael_edgar@milton.edu",
				Payer:            "michael_edgar@milton.edu",
				Payment:          "Credit card",
				PurchaseOrderNum: "N/A",
				Courier:          "FedEx",
				CourierAccount:   "2024-5656-1",
				Comments:         "please include bacterial food source",
				Items:            []string{"DBS0237699", "DBP0000105"},
				ConsumerInfo: &order.UserInfo{
					FirstName:    "Michael",
					LastName:     "Edgar",
					Organization: "Milton Academy",
					FirstAddress: "170 Centre Street",
					City:         "Milton",
					State:        "MA",
					Zipcode:      "02186",
					Country:      "USA",
				},
			},
		},
	}
}

func TestDeterministicLabels(t *testing.T) {
	t.Parallel()

	all := &allData{
		strainData: &strainData{
			strains: []*stock.Strain{{}},
			info: [][]string{
				{"DBS0350001", "GWDI_28_G_7", "gwdi", "GWDI", "blasticidin resistant"},
			},
			invs: [][]string{
				{"GWDI_28_G_7", "axenic cells", "1-2(3)", "3", "pink"},
				{"K. aerogenes GFP", "bacterial stock", "21(17)", "2", "yellow"},
			},
		},
		plasmidData: &plasmidData{
			plasmids: []*stock.Plasmid{{}},
			invs: [][]string{
				{"DBP0001001", "pDGB_A1_0051", "DH5a", "26(9)", "blue"},
			},
		},
	}

	gh := &githubIssue{logger: testLogger()}
	got := gh.resolveLabels(all, sampleOrder())
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

	highConf := 0.95
	fakeClient := &fakeDecisionsClient{
		resp: &components.DecisionsResponse{
			Answers: map[string]components.DecisionsResponseAnswers{
				questionInstitutionType: components.CreateDecisionsResponseAnswersChoice(
					components.DecisionsChoiceAnswer{
						Choice:     labelResearchInst,
						Confidence: &highConf,
						Type:       components.DecisionsChoiceAnswerTypeChoice,
					},
				),
				questionIsMSI: components.CreateDecisionsResponseAnswersNoul(
					components.DecisionsNoulAnswer{
						Noul: 0.91,
						Type: components.DecisionsNoulAnswerTypeNoul,
					},
				),
				questionPaymentLabel: components.CreateDecisionsResponseAnswersChoice(
					components.DecisionsChoiceAnswer{
						Choice:     labelCreditCard,
						Confidence: &highConf,
						Type:       components.DecisionsChoiceAnswerTypeChoice,
					},
				),
				questionLocationLabel: components.CreateDecisionsResponseAnswersChoice(
					components.DecisionsChoiceAnswer{
						Choice:     "MA",
						Confidence: &highConf,
						Type:       components.DecisionsChoiceAnswerTypeChoice,
					},
				),
				questionBacterialFood: components.CreateDecisionsResponseAnswersNoul(
					components.DecisionsNoulAnswer{
						Noul: 0.88,
						Type: components.DecisionsNoulAnswerTypeNoul,
					},
				),
				questionGoldenBraid: components.CreateDecisionsResponseAnswersNoul(
					components.DecisionsNoulAnswer{
						Noul: 0.10,
						Type: components.DecisionsNoulAnswerTypeNoul,
					},
				),
				questionGWDI: components.CreateDecisionsResponseAnswersNoul(
					components.DecisionsNoulAnswer{
						Noul: 0.05,
						Type: components.DecisionsNoulAnswerTypeNoul,
					},
				),
				questionAxenicMedium: components.CreateDecisionsResponseAnswersNoul(
					components.DecisionsNoulAnswer{
						Noul: 0.20,
						Type: components.DecisionsNoulAnswerTypeNoul,
					},
				),
			},
		},
	}

	all := &allData{
		strainData: &strainData{
			strains: []*stock.Strain{{}},
			info:    [][]string{{"DBS0237699", testStrainName, testStrainName, testStrainName, ""}},
			invs:    [][]string{{testStrainName, "bacterial stock", "21(17)", "2", "yellow"}},
		},
		plasmidData: &plasmidData{},
	}

	gh := &githubIssue{
		logger:    testLogger(),
		decisions: fakeClient,
		jevModel:  defaultJevModel,
	}

	got := gh.resolveLabels(all, sampleOrder())

	require.Equal(t, 1, fakeClient.calls)
	require.Equal(t, defaultJevModel, fakeClient.last.Model)
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

	lowConf := 0.40
	lowClient := &fakeDecisionsClient{
		resp: &components.DecisionsResponse{
			Answers: map[string]components.DecisionsResponseAnswers{
				questionInstitutionType: components.CreateDecisionsResponseAnswersChoice(
					components.DecisionsChoiceAnswer{
						Choice:     labelHighSchool,
						Confidence: &lowConf,
						Type:       components.DecisionsChoiceAnswerTypeChoice,
					},
				),
				questionLocationLabel: components.CreateDecisionsResponseAnswersChoice(
					components.DecisionsChoiceAnswer{
						Choice:     choiceNone,
						Confidence: new(0.99),
						Type:       components.DecisionsChoiceAnswerTypeChoice,
					},
				),
			},
		},
	}

	all := &allData{
		strainData:  &strainData{},
		plasmidData: &plasmidData{plasmids: []*stock.Plasmid{{}}},
	}

	ghLow := &githubIssue{
		logger:    testLogger(),
		decisions: lowClient,
		jevModel:  defaultJevModel,
	}
	require.Equal(t, []string{labelPlasmidOrder}, ghLow.resolveLabels(all, sampleOrder()))

	errClient := &fakeDecisionsClient{err: errors.New("openrouter timeout")}
	ghErr := &githubIssue{
		logger:    testLogger(),
		decisions: errClient,
		jevModel:  defaultJevModel,
	}
	require.Equal(t, []string{labelPlasmidOrder}, ghErr.resolveLabels(all, sampleOrder()))
}
