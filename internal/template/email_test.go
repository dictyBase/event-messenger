package template

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func TestEmailPlasmidHtml(t *testing.T) {
	ec := fakePlasmidOnlyEmailContent()
	b, err := OutputHTML(&OutputParams{
		File:    "email.tmpl",
		Path:    "./../../assets",
		Content: ec,
	})
	assert := assert.New(t)
	assert.NoError(err, "expect no error from rendering email template with plasmid data")
	doc, err := goquery.NewDocumentFromReader(b)
	assert.NoError(err, "expect no error from reading html output")
	testOrderHeader(t, doc, ec)
	testOrderAddress(t, doc, ec)
	testOrderPayment(t, doc, ec)
	testOrderPayPlasmid(t, doc, ec)
	testPlasmidInfo(t, doc)
}

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
	testStrainInfo(t, doc)
}

func testPlasmidInfo(t *testing.T, doc *goquery.Document) {
	assert := assert.New(t)
	assert.Exactly(
		doc.Find("div#plasmid.card-panel>h5.blue-text").Text(),
		"Plasmid Information",
		"should match the plasmid information header",
	)
	th := doc.Find(
		"div#plasmid.card-panel>div.section>table.striped>thead>tr",
	).Children().Map(childrenContent)
	assert.Lenf(th, 3, "expect %d got %d elements", 3, len(th))
	assert.ElementsMatch(
		th,
		[]string{"ID", "Plasmid Name", "Citation"},
		"should match all header elements",
	)
	tr := doc.Find(
		"div#plasmid.card-panel>div.section>table.striped>tbody",
	).Children()
	assert.Exactly(tr.Length(), 3, "should have 3 table rows")
	assert.Exactly(tr.Children().Length(), 9, "should have total of 9 columns")
	stItems := fakePlasmidItems()
	tr.Each(func(idx int, sel *goquery.Selection) {
		assert.Exactly(
			sel.Find("td:first-child").Text(),
			stItems[idx],
			"should match the plasmid Id",
		)
		assert.Exactly(
			sel.Find("td:nth-child(2)").Text(),
			"pDV-fAR1-CYFP",
			"should match the plasmid name",
		)
		testPubInfo(assert, sel)
	})
}

func testPubInfo(assert *assert.Assertions, sel *goquery.Selection) {
	assert.Exactly(
		sel.Find("td:last-child>a:first-child").Text(),
		"Pubmed",
		"should match text of first link",
	)
	pubHref, _ := sel.Find("td:last-child>a:first-child").Attr("href")
	assert.Exactly(pubHref, "https://pubmed.gov/26088819", "should match pubmed url")
	assert.Exactly(
		sel.Find("td:last-child>a:nth-child(2)").Text(),
		"Full text",
		"should match text of last link",
	)
	doiHref, _ := sel.Find("td:last-child>a:nth-child(2)").Attr("href")
	assert.Exactly(doiHref, "https://doi.org/10.1002/dvg.22867", "should match doi url")
}

func testStrainInfo(t *testing.T, doc *goquery.Document) {
	assert := assert.New(t)
	assert.Exactly(
		doc.Find("div#strain.card-panel>h5.blue-text").Text(),
		"Strain Information",
		"should match the strain information header",
	)
	th := doc.Find(
		"div#strain.card-panel>div.section>table.striped>thead>tr",
	).Children().Map(childrenContent)
	assert.Lenf(th, 5, "expect %d got %d elements", 5, len(th))
	assert.ElementsMatch(
		th,
		[]string{"ID", "Descriptor", "Name(s)", "Systematic Name", "Citation"},
		"should match all header elements",
	)
	tr := doc.Find(
		"div#strain.card-panel>div.section>table.striped>tbody",
	).Children()
	assert.Exactly(tr.Length(), 4, "should have 4 table rows")
	assert.Exactly(tr.Children().Length(), 20, "should have total of 20 columns")
	stItems := fakeStrainItems()
	tr.Each(func(idx int, sel *goquery.Selection) {
		assert.Exactly(
			sel.Find("td:first-child").Text(),
			stItems[idx],
			"should match the strain Id",
		)
		assert.Exactly(
			sel.Find("td:nth-child(2)").Text(),
			"JB10",
			"should match the strain systematic name",
		)
		assert.Exactly(
			sel.Find("td:nth-child(3)").Text(),
			"jb10ale<br/>jb10 ale<br/>jb10-ale",
			"should match the strain name",
		)
		assert.Exactly(
			sel.Find("td:nth-child(4)").Text(),
			"gefA-",
			"should match the strain descriptor",
		)
		testPubInfo(assert, sel)
	})
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
		fmt.Sprintf("Order #%s", orderID),
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
	pdiv := doc.Find(
		"div#payment-info.card-panel>div.section",
	)
	assert.Regexp(
		regexp.MustCompile("Payment information"),
		pdiv.Text(),
		"should match payment information text",
	)
	assert.Exactly(
		pdiv.Find("a.blue-text.text-darken-1").Text(),
		"DSC website",
		"should match the text for the link",
	)
}

func testOrderPayPlasmid(t *testing.T, doc *goquery.Document, ec *EmailContent) {
	assert := assert.New(t)
	tds := doc.Find(
		"div#cost.card-panel>div.section>table.striped>tbody>tr:nth-child(1)",
	).Children().Map(childrenContent)
	assert.Lenf(tds, 4, "expect %d got %d elements", 4, len(tds))
	assert.ElementsMatch(
		tds,
		[]string{
			"Plasmid",
			strconv.Itoa(ec.PlasmidItems()),
			strconv.Itoa(ec.PlasmidPrice),
			strconv.Itoa(ec.PlasmidCost()),
		},
		"should match all plasmid cost elements",
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
