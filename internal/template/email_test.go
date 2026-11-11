package template

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/dictyBase/event-messenger/internal/datasource"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/johnfercher/go-tree/node"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type textEntry struct {
	text string
	href string
}

func collectTextEntries(n *node.Node[core.Structure]) []textEntry {
	if n == nil {
		return nil
	}

	var out []textEntry

	data := n.GetData()
	if data.Type == "text" {
		val, _ := data.Value.(string)
		href, _ := data.Details["prop_hyperlink"].(string)
		out = append(out, textEntry{text: val, href: href})
	}

	for _, child := range n.GetNexts() {
		out = append(out, collectTextEntries(child)...)
	}

	return out
}

func entryTexts(entries []textEntry) []string {
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		out = append(out, entry.text)
	}

	return out
}

func TestEmailPlasmidPDF(t *testing.T) {
	t.Parallel()

	ec := fakePlasmidOnlyEmailContent()
	m := BuildOrderMaroto(ec)
	entries := collectTextEntries(m.GetStructure())
	texts := entryTexts(entries)

	assertPDFOrderHeader(t, texts, ec)
	assertPDFOrderAddress(t, entries, texts, ec)
	assertPDFOrderPayment(t, entries, texts, ec)
	assertPDFOrderPayPlasmid(t, texts, ec)
	assertPDFPlasmidInfo(t, entries, texts)

	buf, err := OutputPDF(&OutputParams{Content: ec})
	require.NoError(t, err)
	require.True(t, bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")))
}

func TestEmailStrainPDF(t *testing.T) {
	t.Parallel()

	ec := fakeStrainOnlyEmailContent()
	m := BuildOrderMaroto(ec)
	entries := collectTextEntries(m.GetStructure())
	texts := entryTexts(entries)

	assertPDFOrderHeader(t, texts, ec)
	assertPDFOrderAddress(t, entries, texts, ec)
	assertPDFOrderPayment(t, entries, texts, ec)
	assertPDFOrderPayStrain(t, texts, ec)
	assertPDFStrainInfo(t, entries, texts)

	buf, err := OutputPDF(&OutputParams{Content: ec})
	require.NoError(t, err)
	require.True(t, bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")))
}

func TestOutputPDFInvalidContent(t *testing.T) {
	t.Parallel()

	_, err := OutputPDF(&OutputParams{Content: "invalid"})
	require.Error(t, err)
}

func assertPDFOrderHeader(t *testing.T, texts []string, ec *EmailContent) {
	t.Helper()

	require.Contains(t, texts, "Dicty Stock Center")
	require.Contains(t, texts, "Order")
	require.Contains(t, texts, "Confirmation")
	require.Contains(t, texts, fmt.Sprintf("Order #%s", orderID))
	require.Contains(t, texts, ec.OrderTimestamp())
}

func assertPDFOrderAddress(
	t *testing.T,
	entries []textEntry,
	texts []string,
	ec *EmailContent,
) {
	t.Helper()

	require.Contains(t, texts, "Shipping and Billing")
	require.Contains(t, texts, "Information")
	require.Contains(t, texts, "Shipping Address")
	require.Contains(t, texts, "Billing Address")
	require.Contains(t, texts, "Harrold Pennypacker")
	require.Contains(t, texts, "Kel Varnsen")
	require.Contains(
		t,
		texts,
		fmt.Sprintf(
			"%s %s",
			ec.Order.GetData().GetAttributes().GetCourier(),
			ec.Order.GetData().GetAttributes().GetCourierAccount(),
		),
	)
	require.Contains(
		t,
		texts,
		fmt.Sprintf(
			"%s %s",
			ec.Order.GetData().GetAttributes().GetPayment(),
			ec.Order.GetData().GetAttributes().GetPurchaseOrderNum(),
		),
	)
	require.Contains(t, entries, textEntry{
		text: consumer,
		href: "mailto:" + consumer,
	})
	require.Contains(t, entries, textEntry{
		text: payer,
		href: "mailto:" + payer,
	})
	require.Contains(t, texts, "Comment")
	require.Contains(t, texts, ec.Order.GetData().GetAttributes().GetComments())
}

func assertPDFOrderPayment(
	t *testing.T,
	entries []textEntry,
	texts []string,
	ec *EmailContent,
) {
	t.Helper()

	require.Contains(t, texts, "Stocks Ordered")
	require.Contains(t, texts, itemColumn)
	require.Contains(t, texts, quantityColumn)
	require.Contains(t, texts, "Unit Price ($)")
	require.Contains(t, texts, "Total ($)")
	require.Contains(t, texts, totalColumn)
	require.Contains(t, texts, strconv.Itoa(ec.TotalCost()))
	require.Contains(t, texts, "Order Information")
	require.Contains(t, entries, textEntry{
		text: "Order information",
		href: "https://dictycr.org/stockcenter/information/order/",
	})
	require.Contains(t, entries, textEntry{
		text: "Order payment information",
		href: "https://dictycr.org/stockcenter/information/payment",
	})
}

func assertPDFOrderPayPlasmid(t *testing.T, texts []string, ec *EmailContent) {
	t.Helper()

	require.Contains(t, texts, "Plasmid")
	require.Contains(t, texts, strconv.Itoa(ec.PlasmidItems()))
	require.Contains(t, texts, strconv.Itoa(ec.PlasmidPrice))
	require.Contains(t, texts, strconv.Itoa(ec.PlasmidCost()))
}

func assertPDFOrderPayStrain(t *testing.T, texts []string, ec *EmailContent) {
	t.Helper()

	require.Contains(t, texts, "Strain")
	require.Contains(t, texts, strconv.Itoa(ec.StrainItems()))
	require.Contains(t, texts, strconv.Itoa(ec.StrainPrice))
	require.Contains(t, texts, strconv.Itoa(ec.StrainCost()))
}

func assertPDFPlasmidInfo(
	t *testing.T,
	entries []textEntry,
	texts []string,
) {
	t.Helper()

	require.Contains(t, texts, "Plasmid Information")

	for _, id := range fakePlasmidItems() {
		require.Contains(t, texts, id)
	}

	require.Contains(t, texts, plasmidName)
	require.Contains(
		t,
		texts,
		"Thank you for citing the researchers who submitted the materials you ordered!",
	)
	assertPDFPubInfo(t, entries, texts)
}

func assertPDFStrainInfo(
	t *testing.T,
	entries []textEntry,
	texts []string,
) {
	t.Helper()

	require.Contains(t, texts, "Strain Information")

	for _, id := range fakeStrainItems() {
		require.Contains(t, texts, id)
	}

	require.Contains(t, texts, "gefA-")
	require.Contains(t, texts, "JB10")
	require.Contains(t, texts, "jb10ale, jb10 ale, jb10-ale")

	for _, txt := range texts {
		require.NotContains(t, txt, "<br/>")
	}

	assertPDFPubInfo(t, entries, texts)
}

func assertPDFPubInfo(t *testing.T, entries []textEntry, texts []string) {
	t.Helper()

	require.Contains(t, texts, "Basu S et al. (2015)")
	require.Contains(t, entries, textEntry{
		text: "Pubmed",
		href: "https://pubmed.gov/26088819",
	})
	require.Contains(t, entries, textEntry{
		text: "Full text",
		href: "https://doi.org/10.1002/dvg.22867",
	})
}

func indigoDossierFixture() *EmailContent {
	const firstStrainID = "DBS0236922"

	created := time.Date(2026, time.November, 11, 12, 0, 0, 0, time.UTC)

	return &EmailContent{
		Content: &Content{
			StrainPrice:  30,
			PlasmidPrice: 15,
			Shipper: &order.UserInfo{
				FirstName:    "Test",
				LastName:     "Consumer",
				Organization: "Test Lab",
				FirstAddress: "123 Test Street",
				City:         "Testville",
				State:        "TS",
				Zipcode:      "00000",
				Country:      "USA",
				Phone:        "555-0100",
			},
			Payer: &order.UserInfo{
				FirstName:    "Test",
				LastName:     "Payer",
				Organization: "Test Lab",
				FirstAddress: "123 Test Street",
				City:         "Testville",
				State:        "TS",
				Zipcode:      "00000",
				Country:      "USA",
				Phone:        "555-0200",
			},
			Order: &order.Order{
				Data: &order.Order_Data{
					Id: "57081346",
					Attributes: &order.OrderAttributes{
						Consumer:         "basusiddhartha@gmail.com",
						Payer:            "siddhartha-basu@northwestern.edu",
						Courier:          "FedEx",
						CourierAccount:   "FAKE-ACCT-0001",
						Payment:          "credit_card",
						PurchaseOrderNum: "PO-FAKE-0001",
						Comments:         "automated CLI test order",
						CreatedAt:        timestamppb.New(created),
						Items: []string{
							firstStrainID,
							"DBS0391519",
							"DBS0391518",
						},
					},
				},
			},
		},
		StrainData: []*StrainRows{
			{
				ID:         firstStrainID,
				Descriptor: firstStrainID,
				Names:      "sadA-gfp",
				SysName:    "sadA-/[act15]:sadA:GFP",
				PubInfo: []*datasource.PubInfo{
					{
						AuthorStr: "Petra Fey et al. (2002)",
						PubmedURL: "https://pubmed.ncbi.nlm.nih.gov/12097910/",
						DoiURL:    "https://doi.org/10.1242/jcs.115.14.2895",
					},
				},
			},
			{
				ID:         "DBS0391519",
				Descriptor: "[rsmJ]-",
				Names:      "",
				SysName:    "[rsmJ]-",
				PubInfo: []*datasource.PubInfo{
					{
						AuthorStr: "Gruenheit N et al. (2019)",
						DoiURL:    testDoiURL,
					},
				},
			},
			{
				ID:         "DBS0391518",
				Descriptor: "[rpa1]-",
				Names:      "",
				SysName:    "[rpa1]-",
				PubInfo: []*datasource.PubInfo{
					{
						AuthorStr: "Gruenheit N et al. (2019)",
						DoiURL:    testDoiURL,
					},
				},
			},
		},
	}
}

func TestIndigoDossierPreviewPDF(t *testing.T) {
	t.Parallel()

	buf, err := OutputPDF(&OutputParams{Content: indigoDossierFixture()})
	require.NoError(t, err)
	require.True(t, bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")))

	if os.Getenv("WRITE_PREVIEW_PDF") != "" {
		require.NoError(
			t,
			os.WriteFile("testdata-indigo-dossier-preview.pdf", buf.Bytes(), 0o600),
		)
	}
}
