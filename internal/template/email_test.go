package template

import (
	"fmt"
	"regexp"
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
	assert.Exactly(doc.Find("div.row>div.col.s6>div").Length(), 18, "expect to have 18 children")
	selFirst := doc.Find("div.row>div.col.s6>div:first-child")
	assert.Exactly(selFirst.First().Text(), "Harrold Pennypacker", "expect to matcher the consumers name")
	assert.Exactly(selFirst.Last().Text(), "Kel Varnsen", "expect to matcher the payers name")
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
}
