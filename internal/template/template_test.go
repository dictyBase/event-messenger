package template

import (
	"bytes"
	"io"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

const (
	itemColumn        = "Item"
	quantityColumn    = "Quantity"
	totalColumn       = "Total"
	systematicNameCol = "Systematic Name"
	descriptorColumn  = "Descriptor"
	colorColumn       = "Color"
	issueTemplateFile = "issue.tmpl"
	locationColumn    = "Location"
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

func emailSubstr() []string {
	return []string{
		"dsc-header",
		"shipping-row",
		"Order Confirmation",
		"Order #",
		"Shipping Address",
		"Billing Address",
		itemColumn,
		quantityColumn,
		totalColumn,
		systematicNameCol,
		descriptorColumn,
		"Plasmid Name",
		"Comment",
		"Payment information",
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

func TestReadFromBundle(t *testing.T) {
	t.Parallel()

	str, err := ReadFromBundle("/", "email.tmpl")
	require.NoError(t, err, "expect no error from reading email.tmpl template file")
	checkSubstr(str, emailSubstr(), t)

	str2, err := ReadFromBundle("/", issueTemplateFile)
	require.NoError(t, err, "expect no error from reading issue.tmpl template file")
	checkSubstr(str2, issueSubstr(), t)
}

func TestMarkdownOutput(t *testing.T) {
	t.Parallel()

	b, err := OutputText(&OutputParams{
		File:    "test_markdown.tmpl",
		Path:    "/",
		Content: fakeTemplateData(),
	})
	require.NoError(t, err, "expect no error from reading test_html.tmpl template file")

	var out bytes.Buffer

	md := goldmark.New(goldmark.WithExtensions(extension.GFM))
	err = md.Convert(b.Bytes(), &out)
	require.NoError(t, err, "expect no error from markdown conversion")
	testHTMLtree(require.New(t), &out, "h1")
}

func TestHTMLOutput(t *testing.T) {
	t.Parallel()

	b, err := OutputHTML(&OutputParams{
		File:    "test_html.tmpl",
		Path:    "/",
		Content: fakeTemplateData(),
	})
	require.NoError(t, err, "expect no error from reading test_html.tmpl template file")
	testHTMLtree(require.New(t), b, "h4")
}

func childrenContent(_ int, html *goquery.Selection) string {
	return html.Text()
}

func testHTMLtree(assert *require.Assertions, r io.Reader, tag string) {
	doc, err := goquery.NewDocumentFromReader(r)
	assert.NoError(err, "expect no error from reading html output")
	assert.Exactlyf(
		"Stock information", doc.Find(tag).Text(),
		"expected header %s got %s",
		"Stock information",
		doc.Find(tag).Text(),
	)
	th := doc.Find("thead>tr").Children().Map(childrenContent)
	assert.Lenf(th, 2, "expect %d th elements got %d", 2, len(th))
	assert.Exactly("ID", th[0], "expect first th value to be ID")
	assert.Exactly("Name", th[1], "expect second th value to be Name")

	tr := doc.Find("tbody").Children()
	assert.Exactlyf(2, tr.Size(), "expect %d tr element got %d", 2, tr.Size())
	tr.Find("td:first-child").Each(func(_ int, html *goquery.Selection) {
		assert.Regexp("DBS", html.Text(), "expect the value of first child to match DBS")
	})
	tr.Find("td:last-child").Each(func(_ int, html *goquery.Selection) {
		assert.Regexp("ori", html.Text(), "expect the value of first child to match ori")
	})
}
