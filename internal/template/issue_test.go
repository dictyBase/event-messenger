package template

import (
	"bytes"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

func TestIssuePlasmidMkdown(t *testing.T) {
	assert := assert.New(t)
	ic := fakePlasmidOnlyIssueContent()
	b, err := OutputText(&OutputParams{
		File:    "issue.tmpl",
		Path:    "./../../assets",
		Content: ic,
	})
	assert.NoError(err, "expect no error from rending issue content")
	var out bytes.Buffer
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))
	err = md.Convert(b.Bytes(), &out)
	assert.NoError(err, "expect no error from markdown conversion")
	doc, err := goquery.NewDocumentFromReader(&out)
	assert.NoError(err, "expect no error from reading html output")
	testMarkdownOrderHeader(t, doc, ic)
}

func testMarkdownOrderHeader(t *testing.T, doc *goquery.Document, ic *IssueContent) {
	assert := assert.New(t)
	assert.Exactly(
		ic.OrderTimestamp(),
		strings.TrimSpace(
			doc.Find("p:first-child").Contents().Eq(1).Text(),
		),
		"should match order timestamp",
	)
	assert.Exactly(
		orderID,
		strings.TrimSpace(
			doc.Find("p:nth-child(2)").Contents().Eq(1).Text(),
		),
		"should match order id",
	)

}
