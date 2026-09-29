package mailgun

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	"github.com/dictyBase/event-messenger/internal/datasource"
	"github.com/dictyBase/event-messenger/internal/fake"
	"github.com/dictyBase/event-messenger/internal/template"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/mailgun/mailgun-go/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// fakePublicationSource is a lazy publicationSource stub that records
// every id it is asked for.
type fakePublicationSource struct {
	mu    sync.Mutex
	infos map[string]*datasource.PubInfo
	errs  map[string]error
	calls []string
}

func (f *fakePublicationSource) ParsedInfo(
	id string,
) IOE.IOEither[error, *datasource.PubInfo] {
	return func() E.Either[error, *datasource.PubInfo] {
		f.mu.Lock()
		f.calls = append(f.calls, id)
		err := f.errs[id]
		info := f.infos[id]
		f.mu.Unlock()

		if err != nil {
			return E.Left[*datasource.PubInfo](err)
		}

		return E.Right[error](info)
	}
}

// recordedCalls returns a copy of the ids asked for so far.
func (f *fakePublicationSource) recordedCalls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]string(nil), f.calls...)
}

// pubInfoFor builds a PubInfo whose author string tags its origin.
func pubInfoFor(author string) *datasource.PubInfo {
	return &datasource.PubInfo{AuthorStr: author}
}

// emailerWithPublications builds an emailer with only the publication
// source wired, enough for the publication traversal and row tests.
func emailerWithPublications(fake *fakePublicationSource) *mailgunEmailer {
	return &mailgunEmailer{pub: fake}
}

// fakeStrain builds a strain carrying the given publication ids.
func fakeStrain(pubs ...string) *stock.Strain {
	return &stock.Strain{
		Data: &stock.Strain_Data{
			Attributes: &stock.StrainAttributes{Publications: pubs},
		},
	}
}

// fakePlasmid builds a plasmid carrying the given publication ids.
func fakePlasmid(pubs ...string) *stock.Plasmid {
	return &stock.Plasmid{
		Data: &stock.Plasmid_Data{
			Attributes: &stock.PlasmidAttributes{Publications: pubs},
		},
	}
}

// strainInfoRows builds the four-column strain info returned by the
// annotation source, one row per index.
func strainInfoRows(n int) [][]string {
	rows := make([][]string, 0, n)
	for i := range n {
		rows = append(rows, []string{
			fmt.Sprintf("DBS%05d", i),
			descriptorLabel,
			namesLabel,
			sysNameLabel,
		})
	}

	return rows
}

// plasmidInfoRows builds the two-column plasmid info returned by the
// stock source, one row per index.
func plasmidInfoRows(n int) [][]string {
	rows := make([][]string, 0, n)
	for i := range n {
		rows = append(rows, []string{
			fmt.Sprintf("DBP%05d", i),
			plasmidNameLabel,
		})
	}

	return rows
}

const (
	shipperEmail = "shipper@example.com"
	payerEmail   = "payer@example.com"

	// Message envelope used by the send tests.
	senderAddress     = "orders@example.org"
	senderName        = "Dicty Stock Center"
	ccAddress         = "cc@example.org"
	providerMessageID = "provider-id"

	// Row labels shared by the strain and plasmid invoice fixtures.
	descriptorLabel  = "descriptor"
	namesLabel       = "name"
	sysNameLabel     = "sys-name"
	plasmidNameLabel = "plasmid-name"
)

// fakeStockSource is an in-memory stockSource stub that records what it
// was asked for. Its recorded fields are guarded because orderData
// queries the strains and plasmids concurrently.
type fakeStockSource struct {
	mu               sync.Mutex
	recordedOrder    *order.Order
	recordedPatterns []string
	requestedDBS     []string
	requestedDBP     []string
	strains          []*stock.Strain
	plasmids         []*stock.Plasmid
	plasmidInfo      [][]string
	err              error
}

func (f *fakeStockSource) StocksFromItems(
	ord *order.Order,
	pattern string,
) []string {
	f.mu.Lock()
	f.recordedOrder = ord
	f.recordedPatterns = append(f.recordedPatterns, pattern)
	f.mu.Unlock()

	if pattern == "DBS" {
		return []string{fake.StrainID}
	}

	return []string{fake.PlasmidID}
}

func (f *fakeStockSource) GetStrains(ids []string) ([]*stock.Strain, error) {
	f.mu.Lock()
	f.requestedDBS = append(f.requestedDBS, ids...)
	f.mu.Unlock()

	return f.strains, f.err
}

func (f *fakeStockSource) GetPlasmids(
	ids []string,
) ([]*stock.Plasmid, error) {
	f.mu.Lock()
	f.requestedDBP = append(f.requestedDBP, ids...)
	f.mu.Unlock()

	return f.plasmids, f.err
}

func (f *fakeStockSource) GetBasicPlasmidInfo(
	_ []*stock.Plasmid,
) ([][]string, error) {
	return f.plasmidInfo, f.err
}

// patterns returns a copy of the patterns asked for so far.
func (f *fakeStockSource) patterns() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]string(nil), f.recordedPatterns...)
}

// fakeAnnotationSource is an in-memory annotationSource stub.
type fakeAnnotationSource struct {
	strainInfo [][]string
	err        error
}

func (f *fakeAnnotationSource) GetBasicStrainInfo(
	_ []*stock.Strain,
) ([][]string, error) {
	return f.strainInfo, f.err
}

// fakePDFRenderer is a pdfRenderer stub that never touches wkhtmltopdf.
type fakePDFRenderer struct {
	buf *bytes.Buffer
	err error
}

func (f *fakePDFRenderer) OutputPDF(
	_ *template.OutputParams,
) (*bytes.Buffer, error) {
	return f.buf, f.err
}

// testOrder builds an order carrying one strain and one plasmid item.
func testOrder() *order.Order {
	return &order.Order{
		Data: &order.Order_Data{
			Id: "ORD123",
			Attributes: &order.OrderAttributes{
				Consumer:     shipperEmail,
				Payer:        payerEmail,
				Items:        []string{fake.StrainID, fake.PlasmidID},
				ConsumerInfo: &order.UserInfo{FirstName: "Shipper", LastName: "Person"},
				PayerInfo:    &order.UserInfo{FirstName: "Payer", LastName: "Person"},
			},
		},
	}
}

// discardLogger returns a logrus entry that writes nowhere.
func discardLogger() *logrus.Entry {
	lg := logrus.New()
	lg.SetOutput(io.Discard)

	return logrus.NewEntry(lg)
}

// testEmailer wires every seam of the mailer to a stub.
func testEmailer(
	stk stockSource,
	anno annotationSource,
	pdf pdfRenderer,
	pub publicationSource,
	client mailgunClient,
) *mailgunEmailer {
	return newMailgunEmailerWithDependencies(
		&EmailerParams{
			Sender:       senderAddress,
			SenderName:   senderName,
			EmailCC:      ccAddress,
			StrainPrice:  30,
			PlasmidPrice: 15,
			Logger:       discardLogger(),
		},
		stk,
		anno,
		pdf,
		pub,
		client,
	)
}

func TestOrderData(t *testing.T) {
	t.Parallel()

	ord := testOrder()
	stkFake := &fakeStockSource{
		strains:     []*stock.Strain{fakeStrain()},
		plasmids:    []*stock.Plasmid{fakePlasmid()},
		plasmidInfo: [][]string{{fake.PlasmidID, plasmidNameLabel}},
	}
	annoFake := &fakeAnnotationSource{
		strainInfo: [][]string{{
			fake.StrainID,
			descriptorLabel,
			namesLabel,
			sysNameLabel,
		}},
	}
	res, err := E.UnwrapError(ioeutils.ToEither(
		testEmailer(
			stkFake,
			annoFake,
			&fakePDFRenderer{buf: bytes.NewBufferString("dummy-pdf")},
			&fakePublicationSource{
				infos: map[string]*datasource.PubInfo{},
				errs:  map[string]error{},
			},
			nil,
		).orderData(ord),
	))
	require.NoError(t, err)
	require.Same(t, ord, stkFake.recordedOrder)
	require.ElementsMatch(
		t,
		[]string{"DBS", "DBP"},
		stkFake.patterns(),
		"orderData queries strains and plasmids independently and so may schedule them in either order",
	)
	require.Equal(t, []string{fake.StrainID}, stkFake.requestedDBS)
	require.Equal(t, []string{fake.PlasmidID}, stkFake.requestedDBP)
	require.Same(t, ord, res.Order)
	require.Len(t, res.Strains, 1)
	require.Equal(t, fake.StrainID, res.Strains[0].ID)
	require.Equal(t, descriptorLabel, res.Strains[0].Descriptor)
	require.Equal(t, namesLabel, res.Strains[0].Names)
	require.Equal(t, sysNameLabel, res.Strains[0].SysName)
	require.Len(t, res.Plasmids, 1)
	require.Equal(t, fake.PlasmidID, res.Plasmids[0].ID)
	require.Equal(t, plasmidNameLabel, res.Plasmids[0].Name)
}

func TestEmailBodyHermetic(t *testing.T) {
	t.Parallel()

	ord := testOrder()
	email := testEmailer(
		&fakeStockSource{
			strains:     []*stock.Strain{fakeStrain()},
			plasmids:    []*stock.Plasmid{fakePlasmid()},
			plasmidInfo: [][]string{{fake.PlasmidID, plasmidNameLabel}},
		},
		&fakeAnnotationSource{
			strainInfo: [][]string{{
				fake.StrainID,
				descriptorLabel,
				namesLabel,
				sysNameLabel,
			}},
		},
		&fakePDFRenderer{buf: bytes.NewBufferString("dummy-pdf")},
		&fakePublicationSource{
			infos: map[string]*datasource.PubInfo{},
			errs:  map[string]error{},
		},
		nil,
	)

	pkg, err := E.UnwrapError(ioeutils.ToEither(email.emailBody(ord)))
	require.NoError(t, err)
	require.Equal(t, "dummy-pdf", pkg.Body.String())
	require.Same(t, ord, pkg.Data.Order)
}

// fakeMailgunClient is a mailgunClient stub that counts sends and can
// pre-fill a message up to the recipient limit.
type fakeMailgunClient struct {
	underlying   *mailgun.MailgunImpl
	sendCalls    int
	sendID       string
	sendErr      error
	preloadLimit bool
}

func newFakeMailgunClient(
	sendErr error,
	preloadLimit bool,
) *fakeMailgunClient {
	return &fakeMailgunClient{
		underlying:   mailgun.NewMailgun("example.org", "test-key"),
		sendID:       providerMessageID,
		sendErr:      sendErr,
		preloadLimit: preloadLimit,
	}
}

func (f *fakeMailgunClient) NewMessage(
	from, subject, text string,
	to ...string,
) *mailgun.Message {
	m := f.underlying.NewMessage(from, subject, text, to...)

	if f.preloadLimit {
		for i := range mailgun.MaxNumberOfRecipients {
			_ = m.AddRecipient(fmt.Sprintf("user%d@example.org", i))
		}
	}

	return m
}

func (f *fakeMailgunClient) Send(
	_ context.Context,
	_ *mailgun.Message,
) (string, string, error) {
	f.sendCalls++

	return "ok", f.sendID, f.sendErr
}

// sendingEmailer builds a fully stubbed emailer ready to send a message.
func sendingEmailer(client mailgunClient) *mailgunEmailer {
	return testEmailer(
		&fakeStockSource{
			strains:     []*stock.Strain{fakeStrain()},
			plasmids:    []*stock.Plasmid{fakePlasmid()},
			plasmidInfo: [][]string{{fake.PlasmidID, plasmidNameLabel}},
		},
		&fakeAnnotationSource{
			strainInfo: [][]string{{
				fake.StrainID,
				descriptorLabel,
				namesLabel,
				sysNameLabel,
			}},
		},
		&fakePDFRenderer{buf: bytes.NewBufferString("dummy-pdf")},
		&fakePublicationSource{
			infos: map[string]*datasource.PubInfo{},
			errs:  map[string]error{},
		},
		client,
	)
}

// preparedPackage builds the invoice package the preparation tests send.
func preparedPackage(t *testing.T, email *mailgunEmailer) emailPackage {
	t.Helper()

	pkg, err := E.UnwrapError(ioeutils.ToEither(email.emailBody(testOrder())))
	require.NoError(t, err)

	return pkg
}

func TestPrepareSendDescription(t *testing.T) {
	t.Parallel()

	clientFake := newFakeMailgunClient(nil, false)
	email := sendingEmailer(clientFake)

	desc, err := E.UnwrapError(ioeutils.ToEither(
		email.prepareSendDescription(preparedPackage(t, email)),
	))
	require.NoError(t, err)
	require.Equal(
		t,
		0,
		clientFake.sendCalls,
		"preparing a description must not reach the terminal send",
	)
	require.NotNil(t, desc.message)
	require.Equal(
		t,
		3,
		desc.message.RecipientCount(),
		"audience is the shipper plus the configured cc and the distinct payer",
	)
}

func TestCCAddresses(t *testing.T) {
	t.Parallel()

	samePayer := ccDecision{
		Shipper:    shipperEmail,
		Payer:      shipperEmail,
		Configured: ccAddress,
	}
	distinctPayer := ccDecision{
		Shipper:    shipperEmail,
		Payer:      payerEmail,
		Configured: ccAddress,
	}

	require.Equal(t, []string{ccAddress}, ccAddresses(samePayer))
	require.Equal(
		t,
		[]string{ccAddress, payerEmail},
		ccAddresses(distinctPayer),
	)
}

func TestPrepareSendDescriptionZeroSendOnRecipientLimit(t *testing.T) {
	t.Parallel()

	clientFake := newFakeMailgunClient(nil, true)
	email := sendingEmailer(clientFake)

	_, err := E.UnwrapError(ioeutils.ToEither(
		email.prepareSendDescription(preparedPackage(t, email)),
	))
	require.Error(t, err)
	require.Equal(
		t,
		0,
		clientFake.sendCalls,
		"a recipient failure must never reach the terminal send",
	)
}

func TestSendEmailSendFailure(t *testing.T) {
	t.Parallel()

	clientFake := newFakeMailgunClient(errors.New("smtp down"), false)

	err := sendingEmailer(clientFake).SendEmail(testOrder())
	require.ErrorContains(t, err, "smtp down")
	require.ErrorContains(t, err, "error in sending email")
	require.Equal(t, 1, clientFake.sendCalls)
}

func TestSendEmailSuccess(t *testing.T) {
	t.Parallel()

	clientFake := newFakeMailgunClient(nil, false)

	require.NoError(t, sendingEmailer(clientFake).SendEmail(testOrder()))
	require.Equal(t, 1, clientFake.sendCalls)
}

func TestSendEmailZeroSendOnRecipientLimit(t *testing.T) {
	t.Parallel()

	clientFake := newFakeMailgunClient(nil, true)

	require.Error(t, sendingEmailer(clientFake).SendEmail(testOrder()))
	require.Equal(t, 0, clientFake.sendCalls)
}

func TestPubInfoFiltersAndTrimsIDs(t *testing.T) {
	t.Parallel()

	fake := &fakePublicationSource{
		infos: map[string]*datasource.PubInfo{
			"123": pubInfoFor("first"),
			"456": pubInfoFor("second"),
		},
		errs: map[string]error{},
	}

	res, err := E.UnwrapError(ioeutils.ToEither(
		emailerWithPublications(fake).pubInfo(
			[]string{"  ", "", " 123 ", "456", " \t "},
		),
	))
	require.NoError(t, err)
	require.Equal(t, []string{"123", "456"}, fake.recordedCalls())
	require.Len(t, res, 2)
	require.Equal(t, "first", res[0].AuthorStr)
	require.Equal(t, "second", res[1].AuthorStr)
}

func TestPubInfoPreservesOrder(t *testing.T) {
	t.Parallel()

	fake := &fakePublicationSource{
		infos: map[string]*datasource.PubInfo{
			"p1": pubInfoFor("first"),
			"p2": pubInfoFor("second"),
			"p3": pubInfoFor("third"),
		},
		errs: map[string]error{},
	}

	res, err := E.UnwrapError(ioeutils.ToEither(
		emailerWithPublications(fake).pubInfo([]string{"p1", "p2", "p3"}),
	))
	require.NoError(t, err)
	require.Equal(t, []string{"p1", "p2", "p3"}, fake.recordedCalls())

	authors := make([]string, 0, len(res))
	for _, info := range res {
		authors = append(authors, info.AuthorStr)
	}

	require.Equal(t, []string{"first", "second", "third"}, authors)
}

func TestPubInfoShortCircuitsOnFailure(t *testing.T) {
	t.Parallel()

	fake := &fakePublicationSource{
		infos: map[string]*datasource.PubInfo{
			"p1": pubInfoFor("first"),
			"p3": pubInfoFor("third"),
		},
		errs: map[string]error{"p2": errors.New("boom")},
	}

	_, err := E.UnwrapError(ioeutils.ToEither(
		emailerWithPublications(fake).pubInfo([]string{"p1", "p2", "p3"}),
	))
	require.Error(t, err)
	require.ErrorContains(t, err, "boom")
	require.Equal(
		t,
		[]string{"p1", "p2"},
		fake.recordedCalls(),
		"a failure must stop the traversal before p3",
	)
}

func TestNormalizePublicationIDs(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		[]string{"123", "456"},
		normalizePublicationIDs([]string{" 123 ", "", "456", " \t "}),
	)
	require.Empty(t, normalizePublicationIDs(nil))
}

func TestAddStrainPubEmptyPublications(t *testing.T) {
	t.Parallel()

	fake := &fakePublicationSource{
		infos: map[string]*datasource.PubInfo{},
		errs:  map[string]error{},
	}

	res, err := E.UnwrapError(ioeutils.ToEither(
		emailerWithPublications(fake).addStrainPub(
			strainInfoRows(1),
			[]*stock.Strain{fakeStrain()},
		),
	))
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "DBS00000", res[0].ID)
	require.Empty(t, res[0].PubInfo)
	require.Empty(
		t,
		fake.recordedCalls(),
		"a row without publication ids must not query the source",
	)
}

func TestAddStrainPubEnrichesPublications(t *testing.T) {
	t.Parallel()

	fake := &fakePublicationSource{
		infos: map[string]*datasource.PubInfo{
			"p1": pubInfoFor("first"),
			"p2": pubInfoFor("second"),
		},
		errs: map[string]error{},
	}

	res, err := E.UnwrapError(ioeutils.ToEither(
		emailerWithPublications(fake).addStrainPub(
			strainInfoRows(1),
			[]*stock.Strain{fakeStrain("p1", "p2")},
		),
	))
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, descriptorLabel, res[0].Descriptor)
	require.Equal(t, namesLabel, res[0].Names)
	require.Equal(t, sysNameLabel, res[0].SysName)

	authors := make([]string, 0, len(res[0].PubInfo))
	for _, info := range res[0].PubInfo {
		authors = append(authors, info.AuthorStr)
	}

	require.Equal(t, []string{"first", "second"}, authors)
}

func TestAddStrainPubPreservesOrder(t *testing.T) {
	t.Parallel()

	fake := &fakePublicationSource{
		infos: map[string]*datasource.PubInfo{
			"p1": pubInfoFor("first"),
			"p2": pubInfoFor("second"),
			"p3": pubInfoFor("third"),
		},
		errs: map[string]error{},
	}

	res, err := E.UnwrapError(ioeutils.ToEither(
		emailerWithPublications(fake).addStrainPub(
			strainInfoRows(3),
			[]*stock.Strain{
				fakeStrain("p1"),
				fakeStrain("p2"),
				fakeStrain("p3"),
			},
		),
	))
	require.NoError(t, err)
	require.Len(t, res, 3)
	require.Equal(t, "DBS00000", res[0].ID)
	require.Equal(t, "DBS00001", res[1].ID)
	require.Equal(t, "DBS00002", res[2].ID)
	require.Equal(t, "first", res[0].PubInfo[0].AuthorStr)
	require.Equal(t, "second", res[1].PubInfo[0].AuthorStr)
	require.Equal(t, "third", res[2].PubInfo[0].AuthorStr)
}

func TestAddPlasmidPubEmptyPublications(t *testing.T) {
	t.Parallel()

	fake := &fakePublicationSource{
		infos: map[string]*datasource.PubInfo{},
		errs:  map[string]error{},
	}

	res, err := E.UnwrapError(ioeutils.ToEither(
		emailerWithPublications(fake).addPlasmidPub(
			plasmidInfoRows(1),
			[]*stock.Plasmid{fakePlasmid()},
		),
	))
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "DBP00000", res[0].ID)
	require.Empty(t, res[0].PubInfo)
	require.Empty(
		t,
		fake.recordedCalls(),
		"a row without publication ids must not query the source",
	)
}

func TestAddPlasmidPubEnrichesPublications(t *testing.T) {
	t.Parallel()

	fake := &fakePublicationSource{
		infos: map[string]*datasource.PubInfo{
			"p1": pubInfoFor("first"),
		},
		errs: map[string]error{},
	}

	res, err := E.UnwrapError(ioeutils.ToEither(
		emailerWithPublications(fake).addPlasmidPub(
			plasmidInfoRows(1),
			[]*stock.Plasmid{fakePlasmid("p1")},
		),
	))
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, plasmidNameLabel, res[0].Name)
	require.Len(t, res[0].PubInfo, 1)
	require.Equal(t, "first", res[0].PubInfo[0].AuthorStr)
}
