package template

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

func checkSubstr(str string, slice []string, t *testing.T) {
	assert := assert.New(t)
	for _, s := range slice {
		assert.Truef(
			strings.Contains(str, s),
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
	t.Parallel()
	str, err := ReadFromBundle("/", "email.tmpl")
	assert := assert.New(t)
	assert.NoError(err, "expect no error from reading email.tmpl template file")
	checkSubstr(str, emailSubstr(), t)
	str2, err := ReadFromBundle("/", "issue.tmpl")
	assert.NoError(err, "expect no error from reading issue.tmpl template file")
	checkSubstr(str2, issueSubstr(), t)
}

func TestMarkdownOutput(t *testing.T) {
	t.Parallel()
	b, err := OutputText(&OutputParams{
		File:    "test_markdown.tmpl",
		Path:    "/",
		Content: fakeTemplateData(),
	})
	assert := assert.New(t)
	assert.NoError(err, "expect no error from reading test_html.tmpl template file")
	var out bytes.Buffer
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))
	err = md.Convert(b.Bytes(), &out)
	assert.NoError(err, "expect no error from markdown conversion")
	testHTMLtree(assert, &out, "h1")
}

func TestHTMLOutput(t *testing.T) {
	t.Parallel()
	b, err := OutputHTML(&OutputParams{
		File:    "test_html.tmpl",
		Path:    "/",
		Content: fakeTemplateData(),
	})
	assert := assert.New(t)
	assert.NoError(err, "expect no error from reading test_html.tmpl template file")
	testHTMLtree(assert, b, "h4")
}

func childrenContent(index int, html *goquery.Selection) string {
	return html.Text()
}

func testHTMLtree(assert *assert.Assertions, r io.Reader, tag string) {
	doc, err := goquery.NewDocumentFromReader(r)
	assert.NoError(err, "expect no error from reading html output")
	assert.Exactlyf(
		doc.Find(tag).Text(),
		"Stock information",
		"expected header %s got %s",
		"Stock information",
		doc.Find(tag).Text(),
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
