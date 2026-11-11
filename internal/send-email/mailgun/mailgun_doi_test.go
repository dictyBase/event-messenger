package mailgun

import (
	"bytes"
	"testing"

	"github.com/dictyBase/event-messenger/internal/datasource"
	"github.com/dictyBase/event-messenger/internal/fake"
	"github.com/dictyBase/event-messenger/internal/template"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/stretchr/testify/require"
)

type doiPDFRenderer struct {
	t     *testing.T
	calls int
}

func (r *doiPDFRenderer) OutputPDF(args *template.OutputParams) (*bytes.Buffer, error) {
	r.calls++
	ec, ok := args.Content.(*template.EmailContent)
	require.True(r.t, ok)
	require.NotNil(r.t, ec)
	require.Len(r.t, ec.StrainData, 1)
	require.Len(r.t, ec.StrainData[0].PubInfo, 1)
	require.Equal(r.t, "Smith A (2019)", ec.StrainData[0].PubInfo[0].AuthorStr)
	require.Equal(
		r.t,
		"https://doi.org/10.1101/582072",
		ec.StrainData[0].PubInfo[0].DoiURL,
	)
	require.Empty(r.t, ec.StrainData[0].PubInfo[0].PubmedURL)

	body, err := template.OutputPDF(args)
	require.NoError(r.t, err)
	require.True(r.t, bytes.HasPrefix(body.Bytes(), []byte("%PDF-")))

	return bytes.NewBufferString("doi-invoice-pdf"), nil
}

func TestSendEmailDOIPublication(t *testing.T) {
	t.Parallel()

	stk := new(fakeStockSource)
	stk.strains = []*stock.Strain{fakeStrain("doi:10.1101/582072")}
	stk.plasmids = []*stock.Plasmid{fakePlasmid()}
	stk.plasmidInfo = [][]string{{fake.PlasmidID, plasmidNameLabel}}
	anno := new(fakeAnnotationSource)
	anno.strainInfo = [][]string{{fake.StrainID, descriptorLabel, namesLabel, sysNameLabel}}
	pub := new(datasource.PubInfo)
	pub.AuthorStr = "Smith A (2019)"
	pub.DoiURL = "https://doi.org/10.1101/582072"
	pubs := new(fakePublicationSource)
	pubs.infos = map[string]*datasource.PubInfo{"doi:10.1101/582072": pub}
	pdf := new(doiPDFRenderer)
	pdf.t = t
	client := newFakeMailgunClient(nil, false)
	email := testEmailer(stk, anno, pdf, pubs, client)

	require.NoError(t, email.SendEmail(testOrder()))
	require.Equal(t, 1, pdf.calls)
	require.Equal(t, 1, client.sendCalls)
	require.Equal(t, []string{"doi:10.1101/582072"}, pubs.recordedCalls())
}
