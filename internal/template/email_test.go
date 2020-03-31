package template

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func TestEmailStrainHtml(t *testing.T) {
	ec := fakeStrainOnlyEmailContent()
	b, err := OutputHTML(&OutputParams{
		File:    "email.tmpl",
		Path:    "./../../assets",
		Content: ec,
	})
	assert := assert.New(t)
	assert.NoError(err, "expect no error from rendering email template with strain data")
	doc, err := goquery.NewDocumentFromReader(b)
	assert.NoError(err, "expect no error from reading html output")
	testOrderHeader(t, doc, ec)
	testOrderAddress(t, doc, ec)
	testOrderPayment(t, doc, ec)
	testOrderPayStrain(t, doc, ec)
}

func testOrderHeader(t *testing.T, doc *goquery.Document, ec *EmailContent) {
	assert := assert.New(t)
	assert.Regexpf(
		regexp.MustCompile("Order Confirmation"),
		doc.Find("h4").Text(),
		"expected header to match %s got %s",
		"Order Confirmation",
		doc.Find("h4").Text(),
	)
	assert.Exactly(
		ec.OrderTimestamp(),
		doc.Find("div.col.s12.right-align>p:first-child>strong").Text(),
		"should match order timestamp",
	)
	assert.Exactly(
		fmt.Sprintf("Order #%s", orderId),
		doc.Find("div.col.s12.right-align>p:last-child>strong").Text(),
		"should match order id",
	)
}

func testOrderAddress(t *testing.T, doc *goquery.Document, ec *EmailContent) {
	assert := assert.New(t)
	assert.Exactly(
		"Shipping Address",
		doc.Find("div.shipping-row>div:first-child>h6>strong").Text(),
		"should match shipping header",
	)
	assert.Exactly(
		"Billing Address",
		doc.Find("div.shipping-row>div:last-child>h6>strong").Text(),
		"should match shipping header",
	)
	assert.Exactly(doc.Find(
		"div.row>div.col.s6>div").Length(),
		18, "expect to have 18 children",
	)
	selFirst := doc.Find("div.row>div.col.s6>div:first-child")
	assert.Exactly(
		selFirst.First().Text(),
		"Harrold Pennypacker",
		"expect to matcher the consumers name",
	)
	assert.Exactly(
		selFirst.Last().Text(),
		"Kel Varnsen",
		"expect to matcher the payers name",
	)
	selLast := doc.Find("div.row>div.col.s6>div:last-child")
	assert.Exactly(
		selLast.First().Text(),
		fmt.Sprintf(
			"%s %s",
			ec.Order.Data.Attributes.Courier,
			ec.Order.Data.Attributes.CourierAccount,
		),
		"should match courier information",
	)
	assert.Exactly(
		selLast.Last().Text(),
		ec.Order.Data.Attributes.Payment,
		"should match payment information",
	)
	selHref := doc.Find("div.row>div.col.s6>div>a.blue-text.text-darken-1")
	assert.Exactly(
		selHref.First().Text(),
		consumer,
		"should match consumers email",
	)
	assert.Exactly(
		selHref.Last().Text(),
		payer,
		"should match payers email",
	)
}

func testOrderPayment(t *testing.T, doc *goquery.Document, ec *EmailContent) {
	assert := assert.New(t)
	th := doc.Find(
		"div#cost.card-panel>div.section>table.striped>thead>tr",
	).Children().Map(childrenContent)
	assert.Lenf(th, 4, "expect %d got %d elements", 4, len(th))
	assert.ElementsMatch(
		th,
		[]string{"Item", "Quantity", "Unit Price ($)", "Total ($)"},
		"should match all header elements",
	)
	tdt := doc.Find(
		"div#cost.card-panel>div.section>table.striped>tbody>tr:last-child",
	).Children().Map(childrenContent)
	assert.Lenf(tdt, 4, "expect %d got %d elements", 4, len(tdt))
	assert.Exactly(tdt[0], "Total", "should have total header")
	assert.Exactly(
		tdt[len(tdt)-1],
		strconv.Itoa(ec.TotalCost()),
		"should match the total cost of the order",
	)
}

func testOrderPayStrain(t *testing.T, doc *goquery.Document, ec *EmailContent) {
	assert := assert.New(t)
	tds := doc.Find(
		"div#cost.card-panel>div.section>table.striped>tbody>tr:first-child",
	).Children().Map(childrenContent)
	assert.Lenf(tds, 4, "expect %d got %d elements", 4, len(tds))
	assert.ElementsMatch(
		tds,
		[]string{
			"Strain",
			strconv.Itoa(ec.StrainItems()),
			strconv.Itoa(ec.StrainPrice),
			strconv.Itoa(ec.StrainCost()),
		},
		"should match all strain cost elements",
	)
}
