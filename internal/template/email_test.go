package template

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmailPlasmidHtml(t *testing.T) {
	ec := fakePlasmidOnlyEmailContent()
	b, err := OutputHTML(&OutputParams{
		File:    "email.tmpl",
		Path:    "/",
		Content: ec,
	})
	require.NoError(t, err, "expect no error from rendering email template with plasmid data")
	doc, err := goquery.NewDocumentFromReader(b)
	require.NoError(t, err, "expect no error from reading html output")
	testHTMLOrderHeader(t, doc, ec)
	testHTMLOrderAddress(t, doc, ec)
	testHTMLOrderPayment(t, doc, ec)
	testHTMLOrderPayPlasmid(t, doc, ec)
	testHTMLPlasmidInfo(t, doc)
}

func TestEmailStrainHtml(t *testing.T) {
	ec := fakeStrainOnlyEmailContent()
	b, err := OutputHTML(&OutputParams{
		File:    "email.tmpl",
		Path:    "/",
		Content: ec,
	})
	require.NoError(t, err, "expect no error from rendering email template with strain data")
	doc, err := goquery.NewDocumentFromReader(b)
	require.NoError(t, err, "expect no error from reading html output")
	testHTMLOrderHeader(t, doc, ec)
	testHTMLOrderAddress(t, doc, ec)
	testHTMLOrderPayment(t, doc, ec)
	testHTMLOrderPayStrain(t, doc, ec)
	testHTMLStrainInfo(t, doc)
}

func testHTMLPlasmidInfo(t *testing.T, doc *goquery.Document) {
	assert := assert.New(t)
	assert.Exactly(
		"Plasmid Information", doc.Find("div#plasmid.card-panel>h5.blue-text").Text(),
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
	assert.Exactly(3, tr.Length(), "should have 3 table rows")
	assert.Exactly(9, tr.Children().Length(), "should have total of 9 columns")

	stItems := fakePlasmidItems()

	tr.Each(func(idx int, sel *goquery.Selection) {
		assert.Exactly(
			sel.Find("td:first-child").Text(),
			stItems[idx],
			"should match the plasmid Id",
		)
		assert.Exactly(
			"pDV-fAR1-CYFP", sel.Find("td:nth-child(2)").Text(),
			"should match the plasmid name",
		)
		testHTMLPubInfo(assert, sel)
	})
}

func testHTMLPubInfo(assert *assert.Assertions, sel *goquery.Selection) {
	assert.Exactly(
		"Pubmed", sel.Find("td:last-child>a:first-child").Text(),
		"should match text of first link",
	)
	pubHref, _ := sel.Find("td:last-child>a:first-child").Attr("href")
	assert.Exactly("https://pubmed.gov/26088819", pubHref, "should match pubmed url")
	assert.Exactly(
		"Full text", sel.Find("td:last-child>a:nth-child(2)").Text(),
		"should match text of last link",
	)
	doiHref, _ := sel.Find("td:last-child>a:nth-child(2)").Attr("href")
	assert.Exactly("https://doi.org/10.1002/dvg.22867", doiHref, "should match doi url")
}

func testHTMLStrainInfo(t *testing.T, doc *goquery.Document) {
	assert := assert.New(t)
	assert.Exactly(
		"Strain Information", doc.Find("div#strain.card-panel>h5.blue-text").Text(),
		"should match the strain information header",
	)
	th := doc.Find(
		"div#strain.card-panel>div.section>table.striped>thead>tr",
	).Children().Map(childrenContent)
	assert.Lenf(th, 5, "expect %d got %d elements", 5, len(th))
	assert.ElementsMatch(
		th,
		[]string{"ID", descriptorColumn, "Name(s)", systematicNameCol, "Citation"},
		"should match all header elements",
	)

	tr := doc.Find(
		"div#strain.card-panel>div.section>table.striped>tbody",
	).Children()
	assert.Exactly(4, tr.Length(), "should have 4 table rows")
	assert.Exactly(20, tr.Children().Length(), "should have total of 20 columns")

	stItems := fakeStrainItems()

	tr.Each(func(idx int, sel *goquery.Selection) {
		assert.Exactly(
			sel.Find("td:first-child").Text(),
			stItems[idx],
			"should match the strain Id",
		)
		assert.Exactly(
			"JB10", sel.Find("td:nth-child(2)").Text(),
			"should match the strain systematic name",
		)
		assert.Exactly(
			"jb10ale<br/>jb10 ale<br/>jb10-ale", sel.Find("td:nth-child(3)").Text(),
			"should match the strain name",
		)
		assert.Exactly(
			"gefA-", sel.Find("td:nth-child(4)").Text(),
			"should match the strain descriptor",
		)
		testHTMLPubInfo(assert, sel)
	})
}

func testHTMLOrderHeader(t *testing.T, doc *goquery.Document, ec *EmailContent) {
	assert := assert.New(t)
	assert.Regexpf(
		"Order Confirmation",
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

func testHTMLOrderAddress(t *testing.T, doc *goquery.Document, ec *EmailContent) {
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
	assert.Exactly(18, doc.Find(
		"div.row>div.col.s6>div").Length(), "expect to have 18 children",
	)
	selFirst := doc.Find("div.row>div.col.s6>div:first-child")
	assert.Exactly(
		"Harrold Pennypacker", selFirst.First().Text(),
		"expect to matcher the consumers name",
	)
	assert.Exactly(
		"Kel Varnsen", selFirst.Last().Text(),
		"expect to matcher the payers name",
	)

	selLast := doc.Find("div.row>div.col.s6>div:last-child")
	assert.Exactly(
		selLast.First().Text(),
		fmt.Sprintf(
			"%s %s",
			ec.Order.GetData().GetAttributes().GetCourier(),
			ec.Order.GetData().GetAttributes().GetCourierAccount(),
		),
		"should match courier information",
	)
	assert.Exactly(
		selLast.Last().Text(),
		fmt.Sprintf(
			"%s %s",
			ec.Order.GetData().GetAttributes().GetPayment(),
			ec.Order.GetData().GetAttributes().GetPurchaseOrderNum(),
		),
		"should match payment information",
	)

	selHref := doc.Find("div.row>div.col.s6>div>a.blue-text.text-darken-1")
	assert.Exactly(
		consumer, selHref.First().Text(),
		"should match consumers email",
	)
	assert.Exactly(
		payer, selHref.Last().Text(),
		"should match payers email",
	)
}

func testHTMLOrderPayment(t *testing.T, doc *goquery.Document, ec *EmailContent) {
	assert := assert.New(t)
	th := doc.Find(
		"div#cost.card-panel>div.section>table.striped>thead>tr",
	).Children().Map(childrenContent)
	assert.Lenf(th, 4, "expect %d got %d elements", 4, len(th))
	assert.ElementsMatch(
		th,
		[]string{itemColumn, quantityColumn, "Unit Price ($)", "Total ($)"},
		"should match all header elements",
	)

	tdt := doc.Find(
		"div#cost.card-panel>div.section>table.striped>tbody>tr:last-child",
	).Children().Map(childrenContent)
	assert.Lenf(tdt, 4, "expect %d got %d elements", 4, len(tdt))
	assert.Exactly(totalColumn, tdt[0], "should have total header")
	assert.Exactly(
		tdt[len(tdt)-1],
		strconv.Itoa(ec.TotalCost()),
		"should match the total cost of the order",
	)

	pdiv := doc.Find(
		"div#payment-info.card-panel>div.section",
	)
	assert.Regexp(
		"Payment information",
		pdiv.Text(),
		"should match payment information text",
	)
	assert.Exactly(
		"DSC website", pdiv.Find("a.blue-text.text-darken-1").Text(),
		"should match the text for the link",
	)
}

func testHTMLOrderPayPlasmid(t *testing.T, doc *goquery.Document, ec *EmailContent) {
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

func testHTMLOrderPayStrain(t *testing.T, doc *goquery.Document, ec *EmailContent) {
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
