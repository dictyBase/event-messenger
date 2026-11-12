package github

import (
	"context"
	"strings"
	"time"

	A "github.com/IBM/fp-go/v2/array"
	openrouter "github.com/OpenRouterTeam/go-sdk"
	"github.com/OpenRouterTeam/go-sdk/models/components"
	"github.com/OpenRouterTeam/go-sdk/models/operations"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
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

func newDecisionsClient(token string) DecisionsClient {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return nil
	}

	sdk := openrouter.New(openrouter.WithSecurity(trimmed))
	if sdk == nil || sdk.Alpha == nil || sdk.Alpha.Decisions == nil {
		return nil
	}

	return sdk.Alpha.Decisions
}

func resolveJevModel(model string) string {
	trimmed := strings.TrimSpace(model)
	if trimmed == "" {
		return defaultJevModel
	}

	return trimmed
}

func (gh *githubIssue) resolveLabels(all *allData, ord *order.Order) []string {
	base := deterministicLabels(all)
	if gh.decisions == nil {
		return base
	}

	ctx, cancel := context.WithTimeout(context.Background(), jevRequestTimeout)
	defer cancel()

	req := components.DecisionsRequest{
		Model:     resolveJevModel(gh.jevModel),
		State:     components.CreateStateMapOfAny(buildDecisionsState(all, ord)),
		Questions: buildDecisionsQuestions(),
	}

	resp, err := gh.decisions.Create(ctx, req)
	if err != nil {
		if gh.logger != nil {
			gh.logger.Warnf("openrouter jev label classification failed: %v", err)
		}

		return base
	}

	return deduplicateLabels(A.ArrayConcatAll(base, extractJevLabels(resp)))
}

func deterministicLabels(all *allData) []string {
	if all == nil {
		return nil
	}

	var labels []string

	if all.strainData != nil && len(all.strains) > 0 {
		labels = append(labels, labelStrainOrder)
	}

	if all.plasmidData != nil && len(all.plasmids) > 0 {
		labels = append(labels, labelPlasmidOrder)
	}

	if hasAxenicStorage(all) {
		labels = append(labels, labelAxenicMedium)
	}

	if hasGWDIStrains(all) {
		labels = append(labels, labelGWDI)
	}

	if hasBacterialFoodSource(all) {
		labels = append(labels, labelBacterialFood)
	}

	if hasGoldenBraidPlasmids(all) {
		labels = append(labels, labelGoldenBraid)
	}

	return deduplicateLabels(labels)
}

func hasAxenicStorage(all *allData) bool {
	if all.strainData == nil {
		return false
	}

	for _, row := range all.strainData.invs {
		for _, cell := range row {
			lower := strings.ToLower(cell)
			if strings.Contains(lower, "axenic") {
				return true
			}
		}
	}

	return false
}

func hasGWDIStrains(all *allData) bool {
	if all.strainData == nil {
		return false
	}

	for _, row := range A.ArrayConcatAll(all.info, all.strainData.invs) {
		for _, cell := range row {
			if strings.Contains(strings.ToUpper(cell), "GWDI") {
				return true
			}
		}
	}

	return false
}

func hasBacterialFoodSource(all *allData) bool {
	if all.strainData == nil {
		return false
	}

	for _, row := range A.ArrayConcatAll(all.info, all.strainData.invs) {
		for _, cell := range row {
			lower := strings.ToLower(cell)
			if strings.Contains(lower, "bacterial stock") ||
				strings.Contains(lower, "k. aerogenes") ||
				strings.Contains(lower, "klebsiella") {
				return true
			}
		}
	}

	return false
}

func hasGoldenBraidPlasmids(all *allData) bool {
	if all.plasmidData == nil {
		return false
	}

	for _, row := range all.plasmidData.invs {
		for _, cell := range row {
			lower := strings.ToLower(cell)
			if strings.Contains(lower, "pdgb") ||
				strings.Contains(lower, "goldenbraid") {
				return true
			}
		}
	}

	return false
}

func buildDecisionsState(all *allData, ord *order.Order) map[string]any {
	attr := ord.GetData().GetAttributes()
	shipper := attr.GetConsumerInfo()
	payer := attr.GetPayerInfo()

	var (
		strainRows  [][]string
		strainInvs  [][]string
		plasmidInvs [][]string
	)

	if all != nil && all.strainData != nil {
		strainRows = all.info
		strainInvs = all.strainData.invs
	}

	if all != nil && all.plasmidData != nil {
		plasmidInvs = all.plasmidData.invs
	}

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
		"strain_info":       strainRows,
		"strain_inventory":  strainInvs,
		"plasmid_inventory": plasmidInvs,
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
	out := make(map[string]string)
	out[labelCreditCard] = "Credit card or procard payment."
	out[labelPurchaseOrder] = "Purchase order (PO) payment."
	out[labelWireTransfer] = "Wire transfer or bank transfer payment."
	out[labelChartString] = "Northwestern University internal chart string payment."
	out[labelWaiverNoPay] = "Fee waiver requested or no payment."
	out[labelNoCharge] = "Free of charge or replacement order at no charge."
	out[choiceNone] = "Payment method is unspecified."

	return out
}

func choiceQuestion(
	instructions string,
	options map[string]string,
) components.Questions {
	criteria := make(map[string]*components.Criteria, len(options))
	for label, desc := range options {
		criteria[label] = new(components.CreateCriteriaStr(desc))
	}

	return components.CreateQuestionsChoice(
		components.DecisionsChoiceQuestion{
			Criteria: criteria,
			Instructions: components.CreateDecisionsChoiceQuestionInstructionsStr(
				instructions,
			),
			Type: components.DecisionsChoiceQuestionTypeChoice,
		},
	)
}

func noulQuestion(
	instructions string,
	trueDesc string,
	falseDesc string,
) components.Questions {
	return components.CreateQuestionsNoul(
		components.DecisionsNoulQuestion{
			Criteria: &components.DecisionsNoulQuestionCriteria{
				True:  components.CreateTrueStr(trueDesc),
				False: components.CreateFalseStr(falseDesc),
			},
			Instructions: components.CreateDecisionsNoulQuestionInstructionsStr(
				instructions,
			),
			Type: components.DecisionsNoulQuestionTypeNoul,
		},
	)
}

func extractJevLabels(resp *components.DecisionsResponse) []string {
	if resp == nil {
		return nil
	}

	answers := resp.GetAnswers()
	if len(answers) == 0 {
		return nil
	}

	var out []string

	for _, qKey := range []string{
		questionInstitutionType,
		questionPaymentLabel,
		questionLocationLabel,
	} {
		if label := extractChoiceLabel(answers[qKey]); label != "" {
			out = append(out, label)
		}
	}

	noulMappings := []struct {
		key   string
		label string
	}{
		{key: questionIsMSI, label: labelMSI},
		{key: questionBacterialFood, label: labelBacterialFood},
		{key: questionGoldenBraid, label: labelGoldenBraid},
		{key: questionGWDI, label: labelGWDI},
		{key: questionAxenicMedium, label: labelAxenicMedium},
	}

	for _, m := range noulMappings {
		if isHighConfidenceNoul(answers[m.key]) {
			out = append(out, m.label)
		}
	}

	return out
}

func extractChoiceLabel(ans components.DecisionsResponseAnswers) string {
	if ans.DecisionsChoiceAnswer == nil {
		return ""
	}

	choice := strings.TrimSpace(ans.DecisionsChoiceAnswer.GetChoice())
	if choice == "" || choice == choiceNone {
		return ""
	}

	conf := ans.DecisionsChoiceAnswer.GetConfidence()
	if conf != nil && *conf < jevConfidenceThreshold {
		return ""
	}

	if conf == nil {
		probs := ans.DecisionsChoiceAnswer.GetProbabilities()
		if probs[choice] < jevConfidenceThreshold {
			return ""
		}
	}

	return choice
}

func isHighConfidenceNoul(ans components.DecisionsResponseAnswers) bool {
	if ans.DecisionsNoulAnswer == nil {
		return false
	}

	return ans.DecisionsNoulAnswer.GetNoul() >= jevConfidenceThreshold
}

func deduplicateLabels(labels []string) []string {
	seen := make(map[string]struct{}, len(labels))
	out := make([]string, 0, len(labels))

	for _, label := range labels {
		trimmed := strings.TrimSpace(label)
		if trimmed == "" {
			continue
		}

		if _, exists := seen[trimmed]; exists {
			continue
		}

		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}

	return out
}

func locationCriteriaMap() map[string]string {
	criteria := map[string]string{
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

	for code, stateName := range usStateLabels() {
		criteria[code] = "US shipping address in " + stateName + " (" + code + ")."
	}

	return criteria
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
