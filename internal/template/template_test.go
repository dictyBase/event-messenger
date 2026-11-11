package template

import (
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

const (
	colorColumn    = "Color"
	locationColumn = "Location"
)

func checkSubstr(str string, slice []string, t *testing.T) {
	assert := assert.New(t)
	for _, s := range slice {
		assert.Containsf(
			str, s,
			"expect to have the pattern %s",
			s,
		)
	}
}

func issueSubstr() []string {
	return []string{
		"Shipping address",
		"Billing address",
		itemColumn,
		quantityColumn,
		totalColumn,
		systematicNameCol,
		descriptorColumn,
		"Characteristics",
		"Strain storage",
		locationColumn,
		colorColumn,
		"Plasmid information and storage",
	}
}

func TestIssueTemplate(t *testing.T) {
	t.Parallel()

	checkSubstr(issueTemplate, issueSubstr(), t)
}

func childrenContent(_ int, html *goquery.Selection) string {
	return html.Text()
}
