package mailgun

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	L "github.com/IBM/fp-go/v2/optics/lens"
	P "github.com/IBM/fp-go/v2/predicate"
	S "github.com/IBM/fp-go/v2/string"
	"github.com/dictyBase/event-messenger/internal/datasource"
	emailer "github.com/dictyBase/event-messenger/internal/send-email"
	"github.com/dictyBase/event-messenger/internal/template"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	predarrays "github.com/dictyBase/fp-go-loom/predicate/array"
	predord "github.com/dictyBase/fp-go-loom/predicate/ord"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/dictyBase/go-genproto/dictybaseapis/user"
	"github.com/mailgun/mailgun-go/v3"
	"github.com/sirupsen/logrus"
)

const (
	etext = `
Hi %s %s,

Thank you for your order (ID %s). It has been submitted to the Dicty Stock Center (DSC).

Please check the attached PDF for your invoice.
`
)

// publicationSource resolves PubMed ids into citation snippets.
type publicationSource interface {
	ParsedInfo(pmid string) IOE.IOEither[error, *datasource.PubInfo]
}

// stockSource reads stock entities and their basic invoice row info.
type stockSource interface {
	StocksFromItems(ord *order.Order, pattern string) []string
	GetStrains(ids []string) ([]*stock.Strain, error)
	GetPlasmids(ids []string) ([]*stock.Plasmid, error)
	GetBasicPlasmidInfo(plasmids []*stock.Plasmid) ([][]string, error)
}

// annotationSource reads the basic invoice row info for strains.
type annotationSource interface {
	GetBasicStrainInfo(strains []*stock.Strain) ([][]string, error)
}

// userSource resolves the shipper and payer of an order.
type userSource interface {
	UsersInOrder(ord *order.Order) (map[string]*user.User, error)
}

// pdfRenderer renders the invoice attachment.
type pdfRenderer interface {
	OutputPDF(args *template.OutputParams) (*bytes.Buffer, error)
}

// pdfRendererFunc adapts a plain render function into a pdfRenderer.
type pdfRendererFunc func(args *template.OutputParams) (*bytes.Buffer, error)

// OutputPDF renders the invoice attachment.
func (f pdfRendererFunc) OutputPDF(
	p *template.OutputParams,
) (*bytes.Buffer, error) {
	return f(p)
}

// mailgunClient prepares and sends messages.
type mailgunClient interface {
	NewMessage(from, subject, text string, to ...string) *mailgun.Message
	Send(ctx context.Context, m *mailgun.Message) (string, string, error)
}

// emailData is the value the invoice template renders.
type emailData struct {
	Order    *order.Order
	Strains  []*template.StrainRows
	Plasmids []*template.PlasmidRows
	User     map[string]*user.User
}

// strainState accumulates an order's strains and their basic row info.
type strainState struct {
	Order   *order.Order
	Strains []*stock.Strain
	Info    [][]string
}

// plasmidState accumulates an order's plasmids and their basic row info.
type plasmidState struct {
	Order    *order.Order
	Plasmids []*stock.Plasmid
	Info     [][]string
}

// emailPackage carries the invoice data and its rendered PDF.
type emailPackage struct {
	Data emailData
	Body *bytes.Buffer
}

// messageState carries a half-built message through preparation.
type messageState struct {
	Client       mailgunClient
	Message      *mailgun.Message
	ShipperEmail string
	PayerEmail   string
	CCs          []string
	Filename     string
	Attachment   []byte
}

// sendDescription is the prepared terminal effect: one client and one
// message, with no sequencing left to decide.
type sendDescription struct {
	client  mailgunClient
	message *mailgun.Message
}

type mailgunEmailer struct {
	client    mailgunClient
	logger    *logrus.Entry
	anno      annotationSource
	stk       stockSource
	usr       userSource
	pub       publicationSource
	pdf       pdfRenderer
	strprice  int
	plasprice int
	from      string
	name      string
	cc        string
}

// publicationRowState carries one strain or plasmid row while its
// publication info loads.
type publicationRowState struct {
	Emailer *mailgunEmailer
	Info    []string
	PubIDs  []string
	Pubs    []*datasource.PubInfo
}

type EmailerParams struct {
	Sender       string
	SenderName   string
	Domain       string
	APIKey       string
	EmailCC      string
	StrainPrice  int
	PlasmidPrice int
	Logger       *logrus.Entry
	PubSource    *datasource.Publication
	*datasource.Sources
}

// NewMailgunEmailer wires the production dependencies onto the emailer.
func NewMailgunEmailer(args *EmailerParams) emailer.Handler {
	return newMailgunEmailerWithDependencies(
		args,
		args.StockSource,
		args.AnnoSource,
		args.UserSource,
		pdfRendererFunc(template.OutputPDF),
		args.PubSource,
		getMailgunClient(args.Domain, args.APIKey),
	)
}

// newMailgunEmailerWithDependencies wires every seam explicitly so tests
// can inject stubs.
func newMailgunEmailerWithDependencies(
	args *EmailerParams,
	stk stockSource,
	anno annotationSource,
	usr userSource,
	pdf pdfRenderer,
	pub publicationSource,
	client mailgunClient,
) *mailgunEmailer {
	return &mailgunEmailer{
		name:      args.SenderName,
		from:      args.Sender,
		cc:        args.EmailCC,
		strprice:  args.StrainPrice,
		plasprice: args.PlasmidPrice,
		logger:    args.Logger,
		anno:      anno,
		stk:       stk,
		usr:       usr,
		pub:       pub,
		pdf:       pdf,
		client:    client,
	}
}

// SendEmail builds and sends the invoice email for an order.
func (email *mailgunEmailer) SendEmail(ord *order.Order) error {
	return F.Pipe6(
		IOE.Of[error](ord),
		IOE.Chain(email.emailBody),
		IOE.Chain(email.prepareSendDescription),
		IOE.Chain(executeMailgunSend),
		IOE.ChainFirstIOK[error](email.logMessageSent),
		ioeutils.ToEither[error, string],
		E.ToError[string],
	)
}

// ccDecision carries the addresses that decide the CC list.
type ccDecision struct {
	Shipper    string
	Payer      string
	Configured string
}

// payerDiffers reports whether the payer address differs from the
// shipper address.
func payerDiffers(d ccDecision) bool {
	differsFromShipper := predord.NotEqualStr(d.Shipper)

	return differsFromShipper(d.Payer)
}

// ccWithoutPayer returns only the configured CC address.
func ccWithoutPayer(d ccDecision) []string {
	return []string{d.Configured}
}

// ccWithPayer returns the configured CC address followed by the payer.
func ccWithPayer(d ccDecision) []string {
	return []string{d.Configured, d.Payer}
}

// resolveCCs is the pre-bound branch over the payer-distinct decision.
var resolveCCs = P.Fold(ccWithoutPayer, ccWithPayer)

// ccAddresses lists every CC address for the message.
func ccAddresses(d ccDecision) []string {
	resolve := resolveCCs(payerDiffers)

	return resolve(d)
}

// initMessageState precomputes the message envelope and its addresses.
func (email *mailgunEmailer) initMessageState(pkg emailPackage) messageState {
	ord := pkg.Data.Order
	ordID := ord.GetData().GetId()
	shipper := pkg.Data.User["shipper"].GetData().GetAttributes()
	payer := pkg.Data.User["payer"].GetData().GetAttributes()
	sFirst := shipper.GetFirstName()
	sLast := shipper.GetLastName()
	sEmail := shipper.GetEmail()
	pEmail := payer.GetEmail()
	fromHeader := fmt.Sprintf("%s <%s>", email.name, email.from)
	subject := fmt.Sprintf("Order ID:%s %s %s", ordID, sFirst, sLast)
	text := fmt.Sprintf(etext, sFirst, sLast, ordID)
	msg := email.client.NewMessage(fromHeader, subject, text)

	return messageState{
		Client:       email.client,
		Message:      msg,
		ShipperEmail: sEmail,
		PayerEmail:   pEmail,
		CCs: ccAddresses(ccDecision{
			Shipper:    sEmail,
			Payer:      pEmail,
			Configured: email.cc,
		}),
		Filename:   fmt.Sprintf("invoice-%s.pdf", ordID),
		Attachment: pkg.Body.Bytes(),
	}
}

// rawAddRecipient adds the shipper as a recipient.
func rawAddRecipient(s messageState) (messageState, error) {
	err := s.Message.AddRecipient(s.ShipperEmail)

	return s, err
}

// addRecipient adds the shipper as a recipient, contextually wrapped.
func addRecipient(s messageState) IOE.IOEither[error, messageState] {
	return F.Pipe1(
		IOE.TryCatchError(func() (messageState, error) {
			return rawAddRecipient(s)
		}),
		IOE.MapLeft[messageState](func(err error) error {
			return fmt.Errorf("error adding recipient %s: %w", s.ShipperEmail, err)
		}),
	)
}

// addCC is a consumer sink: it takes one address and returns nothing.
func (s messageState) addCC(cc string) {
	s.Message.AddCC(cc)
}

// addCCs runs the sink over every prepared CC address. The sequential
// traversal is required: the sink mutates one shared *mailgun.Message,
// and IO.TraverseArray applies its elements in parallel.
func addCCs(s messageState) IO.IO[[]IO.Void] {
	addOne := IO.FromConsumer(s.addCC)
	traverse := IO.TraverseArraySeq(addOne)

	return traverse(s.CCs)
}

// attachInvoice attaches the rendered invoice PDF.
func attachInvoice(s messageState) IO.IO[IO.Void] {
	return IO.FromImpure(func() {
		s.Message.AddBufferAttachment(s.Filename, s.Attachment)
	})
}

// toSendDescription projects the prepared state onto the terminal
// effect description.
func toSendDescription(s messageState) sendDescription {
	return sendDescription{client: s.Client, message: s.Message}
}

// prepareSendDescription builds the message without sending it.
func (email *mailgunEmailer) prepareSendDescription(
	pkg emailPackage,
) IOE.IOEither[error, sendDescription] {
	return F.Pipe6(
		pkg,
		email.initMessageState,
		IOE.Of[error],
		IOE.Chain(addRecipient),
		IOE.ChainFirstIOK[error](addCCs),
		IOE.ChainFirstIOK[error](attachInvoice),
		IOE.Map[error](toSendDescription),
	)
}

// rawMailgunSend is the single call site of the terminal mailgun send.
func rawMailgunSend(
	ctx context.Context,
	client mailgunClient,
	msg *mailgun.Message,
) (string, error) {
	_, id, err := client.Send(ctx, msg)

	return id, err
}

// executeMailgunSend performs the terminal send exactly once.
func executeMailgunSend(desc sendDescription) IOE.IOEither[error, string] {
	ctx := context.Background()

	return F.Pipe1(
		IOE.TryCatchError(func() (string, error) {
			return rawMailgunSend(ctx, desc.client, desc.message)
		}),
		IOE.MapLeft[string](func(err error) error {
			return fmt.Errorf("error in sending email: %w", err)
		}),
	)
}

// logMessageSent records the provider message id through the mailer's
// logger. Named IO tap: the logger is a *logrus.Entry on the component,
// and IO.Logf would bypass the configured logrus output.
func (email *mailgunEmailer) logMessageSent(id string) IO.IO[IO.Void] {
	return IO.FromImpure(func() {
		email.logger.Infof("message sent with id %s", id)
	})
}

// fetchStrains loads the strains listed in the order.
func (email *mailgunEmailer) fetchStrains(
	s strainState,
) IOE.IOEither[error, []*stock.Strain] {
	ids := email.stk.StocksFromItems(s.Order, "DBS")

	return F.Pipe1(
		IOE.TryCatchError(func() ([]*stock.Strain, error) {
			return email.stk.GetStrains(ids)
		}),
		IOE.MapLeft[[]*stock.Strain](func(err error) error {
			return fmt.Errorf("error in getting strains: %w", err)
		}),
	)
}

// fetchStrainInfo loads the basic invoice row info of the state's strains.
func (email *mailgunEmailer) fetchStrainInfo(
	s strainState,
) IOE.IOEither[error, [][]string] {
	return F.Pipe1(
		IOE.TryCatchError(func() ([][]string, error) {
			return email.anno.GetBasicStrainInfo(s.Strains)
		}),
		IOE.MapLeft[[][]string](func(err error) error {
			return fmt.Errorf("error in getting strain information: %w", err)
		}),
	)
}

// enrichStrains enriches the state's strain rows with publication info.
func (email *mailgunEmailer) enrichStrains(
	s strainState,
) IOE.IOEither[error, []*template.StrainRows] {
	return email.addStrainPub(s.Info, s.Strains)
}

// strains builds the invoice strain rows for an order.
func (email *mailgunEmailer) strains(
	ord *order.Order,
) IOE.IOEither[error, []*template.StrainRows] {
	return F.Pipe3(
		IOE.Of[error](strainState{Order: ord}),
		IOE.Bind(strainItemsLens.Set, email.fetchStrains),
		IOE.Bind(strainInfoLens.Set, email.fetchStrainInfo),
		IOE.Chain(email.enrichStrains),
	)
}

// fetchPlasmids loads the plasmids listed in the order.
func (email *mailgunEmailer) fetchPlasmids(
	s plasmidState,
) IOE.IOEither[error, []*stock.Plasmid] {
	ids := email.stk.StocksFromItems(s.Order, "DBP")

	return F.Pipe1(
		IOE.TryCatchError(func() ([]*stock.Plasmid, error) {
			return email.stk.GetPlasmids(ids)
		}),
		IOE.MapLeft[[]*stock.Plasmid](func(err error) error {
			return fmt.Errorf("error in getting plasmids: %w", err)
		}),
	)
}

// fetchPlasmidInfo loads the basic invoice row info of the state's
// plasmids.
func (email *mailgunEmailer) fetchPlasmidInfo(
	s plasmidState,
) IOE.IOEither[error, [][]string] {
	return F.Pipe1(
		IOE.TryCatchError(func() ([][]string, error) {
			return email.stk.GetBasicPlasmidInfo(s.Plasmids)
		}),
		IOE.MapLeft[[][]string](func(err error) error {
			return fmt.Errorf("error in getting plasmid information: %w", err)
		}),
	)
}

// enrichPlasmids enriches the state's plasmid rows with publication info.
func (email *mailgunEmailer) enrichPlasmids(
	s plasmidState,
) IOE.IOEither[error, []*template.PlasmidRows] {
	return email.addPlasmidPub(s.Info, s.Plasmids)
}

// plasmids builds the invoice plasmid rows for an order.
func (email *mailgunEmailer) plasmids(
	ord *order.Order,
) IOE.IOEither[error, []*template.PlasmidRows] {
	return F.Pipe3(
		IOE.Of[error](plasmidState{Order: ord}),
		IOE.Bind(plasmidItemsLens.Set, email.fetchPlasmids),
		IOE.Bind(plasmidInfoLens.Set, email.fetchPlasmidInfo),
		IOE.Chain(email.enrichPlasmids),
	)
}

// fetchUsers resolves the shipper and payer of the order.
func (email *mailgunEmailer) fetchUsers(
	ord *order.Order,
) IOE.IOEither[error, map[string]*user.User] {
	return F.Pipe1(
		IOE.TryCatchError(func() (map[string]*user.User, error) {
			return email.usr.UsersInOrder(ord)
		}),
		IOE.MapLeft[map[string]*user.User](func(err error) error {
			return fmt.Errorf("error in getting users for order: %w", err)
		}),
	)
}

// orderData builds every value the invoice needs for an order.
func (email *mailgunEmailer) orderData(
	ord *order.Order,
) IOE.IOEither[error, emailData] {
	return F.Pipe3(
		IOE.Of[error](emailData{Order: ord}),
		IOE.ApS(emailStrainsLens.Set, email.strains(ord)),
		IOE.ApS(emailPlasmidsLens.Set, email.plasmids(ord)),
		IOE.ApS(emailUserLens.Set, email.fetchUsers(ord)),
	)
}

var (
	// strainItemsLens focuses the strains loaded for an order.
	strainItemsLens = L.MakeLens(
		func(s strainState) []*stock.Strain { return s.Strains },
		func(s strainState, v []*stock.Strain) strainState {
			s.Strains = v
			return s
		},
	)

	// strainInfoLens focuses the basic row info of the loaded strains.
	strainInfoLens = L.MakeLens(
		func(s strainState) [][]string { return s.Info },
		func(s strainState, v [][]string) strainState {
			s.Info = v
			return s
		},
	)

	// plasmidItemsLens focuses the plasmids loaded for an order.
	plasmidItemsLens = L.MakeLens(
		func(s plasmidState) []*stock.Plasmid { return s.Plasmids },
		func(s plasmidState, v []*stock.Plasmid) plasmidState {
			s.Plasmids = v
			return s
		},
	)

	// plasmidInfoLens focuses the basic row info of the loaded plasmids.
	plasmidInfoLens = L.MakeLens(
		func(s plasmidState) [][]string { return s.Info },
		func(s plasmidState, v [][]string) plasmidState {
			s.Info = v
			return s
		},
	)

	// emailStrainsLens focuses the invoice's strain rows.
	emailStrainsLens = L.MakeLens(
		func(d emailData) []*template.StrainRows { return d.Strains },
		func(d emailData, v []*template.StrainRows) emailData {
			d.Strains = v
			return d
		},
	)

	// emailPlasmidsLens focuses the invoice's plasmid rows.
	emailPlasmidsLens = L.MakeLens(
		func(d emailData) []*template.PlasmidRows { return d.Plasmids },
		func(d emailData, v []*template.PlasmidRows) emailData {
			d.Plasmids = v
			return d
		},
	)

	// emailUserLens focuses the invoice's resolved users.
	emailUserLens = L.MakeLens(
		func(d emailData) map[string]*user.User { return d.User },
		func(d emailData, v map[string]*user.User) emailData {
			d.User = v
			return d
		},
	)

	// emailBodyLens focuses the rendered invoice PDF.
	emailBodyLens = L.MakeLens(
		func(pkg emailPackage) *bytes.Buffer { return pkg.Body },
		func(pkg emailPackage, v *bytes.Buffer) emailPackage {
			pkg.Body = v
			return pkg
		},
	)
)

// renderPDF renders the invoice PDF for the package's order data.
func (email *mailgunEmailer) renderPDF(
	pkg emailPackage,
) IOE.IOEither[error, *bytes.Buffer] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*bytes.Buffer, error) {
			return email.pdf.OutputPDF(&template.OutputParams{
				Path: "/",
				File: "email.tmpl",
				Content: &template.EmailContent{
					StrainData:  pkg.Data.Strains,
					PlasmidData: pkg.Data.Plasmids,
					Content: &template.Content{
						Order:        pkg.Data.Order,
						Shipper:      pkg.Data.User["shipper"],
						Payer:        pkg.Data.User["payer"],
						StrainPrice:  email.strprice,
						PlasmidPrice: email.plasprice,
					},
				},
			})
		}),
		IOE.MapLeft[*bytes.Buffer](func(err error) error {
			return fmt.Errorf("error in generating invoice pdf: %w", err)
		}),
	)
}

// renderPackage renders the invoice PDF into the package body.
func (email *mailgunEmailer) renderPackage(
	data emailData,
) IOE.IOEither[error, emailPackage] {
	return F.Pipe1(
		IOE.Of[error](emailPackage{Data: data}),
		IOE.Bind(emailBodyLens.Set, email.renderPDF),
	)
}

// emailBody builds the invoice package for an order.
func (email *mailgunEmailer) emailBody(
	ord *order.Order,
) IOE.IOEither[error, emailPackage] {
	return F.Pipe2(
		ord,
		email.orderData,
		IOE.Chain(email.renderPackage),
	)
}

// publicationRowPubsLens focuses the loaded publication info of a row.
var publicationRowPubsLens = L.MakeLens(
	func(s publicationRowState) []*datasource.PubInfo { return s.Pubs },
	func(s publicationRowState, v []*datasource.PubInfo) publicationRowState {
		s.Pubs = v
		return s
	},
)

// publicationRowPubIDs extracts the row's normalized publication ids.
func publicationRowPubIDs(s publicationRowState) []string {
	return s.PubIDs
}

// rowHasPubs reports whether the row carries any publication ids.
var rowHasPubs = F.Pipe1(
	predarrays.IsNonEmpty[string](),
	P.ContraMap(publicationRowPubIDs),
)

// fetchPubs loads publication info for the row's ids.
func (s publicationRowState) fetchPubs() IOE.IOEither[error, []*datasource.PubInfo] {
	return s.Emailer.pubInfo(s.PubIDs)
}

// keepRow leaves a row without publications unchanged.
func keepRow(s publicationRowState) IOE.IOEither[error, publicationRowState] {
	return IOE.Of[error](s)
}

// loadRowPubs enriches a row with the publication info for its ids.
func loadRowPubs(s publicationRowState) IOE.IOEither[error, publicationRowState] {
	return F.Pipe1(
		IOE.Of[error](s),
		IOE.Bind(publicationRowPubsLens.Set, publicationRowState.fetchPubs),
	)
}

// resolveRowPubs is the pre-bound branch: enrich only when ids exist.
var resolveRowPubs = P.Fold(keepRow, loadRowPubs)

// enrichRow applies the pre-bound branch to the row.
func enrichRow(s publicationRowState) IOE.IOEither[error, publicationRowState] {
	resolve := resolveRowPubs(rowHasPubs)

	return F.Pipe1(s, resolve)
}

// toStrainRow projects a row state onto a strain row.
func toStrainRow(s publicationRowState) *template.StrainRows {
	return &template.StrainRows{
		ID:         s.Info[0],
		Descriptor: s.Info[1],
		Names:      s.Info[2],
		SysName:    s.Info[3],
		PubInfo:    s.Pubs,
	}
}

// toPlasmidRow projects a row state onto a plasmid row.
func toPlasmidRow(s publicationRowState) *template.PlasmidRows {
	return &template.PlasmidRows{
		ID:      s.Info[0],
		Name:    s.Info[1],
		PubInfo: s.Pubs,
	}
}

// toStrainRowIO enriches one row and projects it onto a strain row.
func toStrainRowIO(
	s publicationRowState,
) IOE.IOEither[error, *template.StrainRows] {
	return F.Pipe2(
		s,
		enrichRow,
		IOE.Map[error](toStrainRow),
	)
}

// toPlasmidRowIO enriches one row and projects it onto a plasmid row.
func toPlasmidRowIO(
	s publicationRowState,
) IOE.IOEither[error, *template.PlasmidRows] {
	return F.Pipe2(
		s,
		enrichRow,
		IOE.Map[error](toPlasmidRow),
	)
}

// rowBuilder pairs basic row info with the protobuf stock entities.
type rowBuilder struct {
	Emailer *mailgunEmailer
	Info    [][]string
}

// seedState builds the enrichment state for one row, storing the ids
// already normalized so the branch predicate is exact.
func (b rowBuilder) seedState(i int, pubs []string) publicationRowState {
	return publicationRowState{
		Emailer: b.Emailer,
		Info:    b.Info[i],
		PubIDs:  normalizePublicationIDs(pubs),
	}
}

// resolveStrain enriches and projects one strain row.
func (b rowBuilder) resolveStrain(
	i int,
	str *stock.Strain,
) IOE.IOEither[error, *template.StrainRows] {
	pubs := str.GetData().GetAttributes().GetPublications()
	seed := b.seedState(i, pubs)

	return toStrainRowIO(seed)
}

// resolvePlasmid enriches and projects one plasmid row.
func (b rowBuilder) resolvePlasmid(
	i int,
	pls *stock.Plasmid,
) IOE.IOEither[error, *template.PlasmidRows] {
	pubs := pls.GetData().GetAttributes().GetPublications()
	seed := b.seedState(i, pubs)

	return toPlasmidRowIO(seed)
}

func (email *mailgunEmailer) addPlasmidPub(
	strInfo [][]string,
	plasmids []*stock.Plasmid,
) IOE.IOEither[error, []*template.PlasmidRows] {
	builder := rowBuilder{Emailer: email, Info: strInfo}
	traverse := IOE.TraverseArrayWithIndexSeq(builder.resolvePlasmid)

	return F.Pipe1(plasmids, traverse)
}

func (email *mailgunEmailer) addStrainPub(
	strInfo [][]string,
	strains []*stock.Strain,
) IOE.IOEither[error, []*template.StrainRows] {
	builder := rowBuilder{Emailer: email, Info: strInfo}
	traverse := IOE.TraverseArrayWithIndexSeq(builder.resolveStrain)

	return F.Pipe1(strains, traverse)
}

// loadPubInto loads one publication and appends it to the accumulated
// slice.
func (email *mailgunEmailer) loadPubInto(
	id string,
	pubs []*datasource.PubInfo,
) IOE.IOEither[error, []*datasource.PubInfo] {
	appendTo := F.Curry2(A.Append[*datasource.PubInfo])

	return F.Pipe1(
		email.pub.ParsedInfo(id),
		IOE.Map[error](appendTo(pubs)),
	)
}

// loadPubNext appends one publication to the monadic accumulator.
// Chaining on the accumulator is what stops the fold at the first
// failure, so no later id is requested.
func (email *mailgunEmailer) loadPubNext(
	acc IOE.IOEither[error, []*datasource.PubInfo],
	id string,
) IOE.IOEither[error, []*datasource.PubInfo] {
	load := F.Curry2(email.loadPubInto)
	loadID := load(id)

	return F.Pipe1(acc, IOE.Chain(loadID))
}

// pubInfo loads publication info for every id in order, stopping at the
// first failure.
func (email *mailgunEmailer) pubInfo(
	ids []string,
) IOE.IOEither[error, []*datasource.PubInfo] {
	reduce := A.Reduce(
		email.loadPubNext,
		IOE.Of[error]([]*datasource.PubInfo{}),
	)

	return F.Pipe2(
		ids,
		normalizePublicationIDs,
		reduce,
	)
}

// normalizePublicationIDs trims each id and drops the blank ones.
func normalizePublicationIDs(ids []string) []string {
	return F.Pipe2(
		ids,
		A.Map(strings.TrimSpace),
		A.Filter(S.IsNonEmpty),
	)
}

func getMailgunClient(domain, apiKey string) *mailgun.MailgunImpl {
	return mailgun.NewMailgun(domain, apiKey)
}
