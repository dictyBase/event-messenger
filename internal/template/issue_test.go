package template

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	mr "github.com/yuin/goldmark/renderer/html"
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
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(mr.WithUnsafe()),
	)
	err = md.Convert(b.Bytes(), &out)
	assert.NoError(err, "expect no error from markdown conversion")
	doc, err := goquery.NewDocumentFromReader(&out)
	assert.NoError(err, "expect no error from reading html output")
	testMarkdownOrderHeader(t, doc, ic)
	testMarkdownOrderAddress(t, doc, ic)
	testMarkdownOrderPayment(t, doc, ic)
	testMarkdownOrderPayPlasmid(t, doc, ic)
}

func testMarkdownOrderPayment(t *testing.T, doc *goquery.Document, ic *IssueContent) {
	assert := assert.New(t)
	th := doc.Find("table>thead").Eq(1).Find("tr").
		Children().Map(childrenContent)
	assert.Lenf(th, 4, "expect %d got %d elements", 4, len(th))
	assert.ElementsMatch(
		th,
		[]string{"Item", "Quantity", "Unit price($)", "Total($)"},
		"should match all header elements",
	)
	assert.Exactly(
		doc.Find("table>tbody").Eq(1).
			Find("tr:nth-child(3)>td:nth-child(4)").Text(),
		strconv.Itoa(ic.TotalCost()),
		"should match the total cost of the order",
	)
}

func testMarkdownOrderPayPlasmid(t *testing.T, doc *goquery.Document, ic *IssueContent) {
	assert := assert.New(t)
	assert.ElementsMatch(
		doc.Find("table>tbody").Eq(1).
			Find("tr:nth-child(2)").
			Children().Map(childrenContent),
		[]string{
			"Plasmid",
			strconv.Itoa(ic.PlasmidItems()),
			strconv.Itoa(ic.PlasmidPrice),
			strconv.Itoa(ic.PlasmidCost()),
		},
		"should match plasmid row elements",
	)

	//tdt := doc.Find(
	//"div#cost.card-panel>div.section>table.striped>tbody>tr:last-child",
	//).Children().Map(childrenContent)
	//assert.Lenf(tdt, 4, "expect %d got %d elements", 4, len(tdt))
	//assert.Exactly(tdt[0], "Total", "should have total header")
	//assert.Exactly(
	//tdt[len(tdt)-1],
	//strconv.Itoa(ec.TotalCost()),
	//"should match the total cost of the order",
	//)
	//pdiv := doc.Find(
	//"div#payment-info.card-panel>div.section",
	//)
	//assert.Regexp(
	//regexp.MustCompile("Payment information"),
	//pdiv.Text(),
	//"should match payment information text",
	//)
	//assert.Exactly(
	//pdiv.Find("a.blue-text.text-darken-1").Text(),
	//"DSC website",
	//"should match the text for the link",
	//)
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

func testMarkdownOrderAddress(t *testing.T, doc *goquery.Document, ic *IssueContent) {
	assert := assert.New(t)
	assert.Exactly(
		"Shipping address",
		doc.Find("table>thead").Eq(0).Find("tr>th:first-child").Text(),
		"should match shipping header",
	)
	assert.Exactly(
		"Billing address",
		doc.Find("table>thead").Eq(0).Find("tr>th:nth-child(3)").Text(),
		"should match billing header",
	)
	assert.Exactly(
		"Harrold Pennypacker",
		strings.TrimSpace(
			doc.Find("table>tbody").Eq(0).
				Find("tr>td:nth-child(1)").Contents().
				Eq(0).Text(),
		),
		"expect to matcher the consumers name",
	)
	assert.Exactly(
		"pennpacker@dictycr.org",
		strings.TrimSpace(
			doc.Find("table>tbody").Eq(0).
				Find("tr>td:nth-child(1)>a").Text(),
		),
		"expect to matcher the consumers email",
	)
	assert.Exactly(
		"Kel Varnsen",
		strings.TrimSpace(
			doc.Find("table>tbody").Eq(0).
				Find("tr>td:nth-child(3)").Contents().
				Eq(0).Text(),
		),
		"expect to matcher the payers name",
	)
	assert.Exactly(
		"varnsen@dictycr.org",
		strings.TrimSpace(
			doc.Find("table>tbody").Eq(0).
				Find("tr>td:nth-child(3)>a").Text(),
		),
		"expect to matcher the payers email",
	)
	assert.Exactly(
		strings.TrimPrefix(
			doc.Find("table>tbody").Eq(0).
				Find("tr>td:nth-child(1)").Contents().
				Last().Text(),
			" "),
		fmt.Sprintf(
			"%s %s",
			ic.Order.Data.Attributes.Courier,
			ic.Order.Data.Attributes.CourierAccount,
		),
		"should match courier information",
	)
	assert.Exactly(
		strings.TrimPrefix(
			doc.Find("table>tbody").Eq(0).
				Find("tr>td:nth-child(3)").Contents().
				Last().Text(),
			" "),
		ic.Order.Data.Attributes.Payment,
		"should match payment information",
	)
}
