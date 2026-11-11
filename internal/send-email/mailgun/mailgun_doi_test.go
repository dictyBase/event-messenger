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
	body, err := template.OutputHTML(args)
	require.NoError(r.t, err)
	require.Contains(r.t, body.String(), "https://doi.org/10.1101/582072")
	require.Contains(r.t, body.String(), "Smith A (2019)")
	require.NotContains(r.t, body.String(), `href=""`)

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
