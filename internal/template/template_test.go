package template

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func checkSubstr(b fmt.Stringer, str []string, t *testing.T) {
	assert := assert.New(t)
	for _, s := range str {
		assert.Truef(
			strings.Contains(b.String(), s),
			"expect to have the pattern %s",
			s,
		)
	}
}

func emailSubstr() []string {
	return []string{
		"dsc-header",
		"shipping-row",
		"Order Confirmation",
		"Order #",
		"Shipping Address",
		"Billing Address",
		"Item",
		"Quantity",
		"Total",
		"Systematic Name",
		"Descriptor",
		"Plasmid Name",
		"Comment",
		"Payment information",
	}
}

func issueSubstr() []string {
	return []string{
		"Shipping address",
		"Billing address",
		"Item",
		"Quantity",
		"Total",
		"Systematic Name",
		"Descriptor",
		"Characteristics",
		"Strain storage",
		"Location",
		"Color",
		"Plasmid information and storage",
	}
}

func TestReadFromBundle(t *testing.T) {
	b, err := ReadFromBundle("./../../assets", "email.tmpl")
	assert := assert.New(t)
	assert.NoError(err, "expect no error from reading email.tmpl template file")
	checkSubstr(b, emailSubstr(), t)
	b2, err := ReadFromBundle("./../../assets", "issue.tmpl")
	assert.NoError(err, "expect no error from reading issue.tmpl template file")
	checkSubstr(b2, issueSubstr(), t)
}

func TestOutputHTML(t *testing.T) {
	st := []struct {
		ID   string
		Name string
	}{
		{"DBS0236831", "tori"},
		{"DBS0236415", "lori"},
	}
	ct := struct {
		Header  string
		Strains []struct {
			ID   string
			Name string
		}
	}{
		"Stock information",
		st,
	}
	b, err := OutputHTML(&OutputParams{
		File:    "test_html.tmpl",
		Path:    "./../../testdata",
		Content: ct,
	})
	assert := assert.New(t)
	assert.NoError(err, "expect no error from reading test_html.tmpl template file")
	doc, err := goquery.NewDocumentFromReader(b)
	assert.NoError(err, "expect no error from reading html output")
	assert.Exactlyf(
		doc.Find("h4").Text(),
		"Stock information",
		"expected header %s got %s",
		"Stock information",
		doc.Find("h4").Text(),
	)
	th := doc.Find("thead>tr").Children().Map(childrenContent)
	assert.Lenf(th, 2, "expect %d th elements got %d", 2, len(th))
	assert.Exactly(th[0], "ID", "expect first th value to be ID")
	assert.Exactly(th[1], "Name", "expect second th value to be Name")
	tr := doc.Find("tbody").Children()
	assert.Exactlyf(tr.Size(), 2, "expect %d tr element got %d", 2, tr.Size())
	tr.Find("td:first-child").Each(func(i int, html *goquery.Selection) {
		assert.Regexp(regexp.MustCompile("DBS"), html.Text(), "expect the value of first child to match DBS")
	})
	tr.Find("td:last-child").Each(func(i int, html *goquery.Selection) {
		assert.Regexp(regexp.MustCompile("ori"), html.Text(), "expect the value of first child to match ori")
	})
}

func childrenContent(index int, html *goquery.Selection) string {
	return html.Text()
}
