package mailgun

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	L "github.com/IBM/fp-go/v2/optics/lens"
	P "github.com/IBM/fp-go/v2/predicate"
	S "github.com/IBM/fp-go/v2/string"
	"github.com/dictyBase/event-messenger/internal/datasource"
	emailer "github.com/dictyBase/event-messenger/internal/send-email"
	"github.com/dictyBase/event-messenger/internal/template"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	predarrays "github.com/dictyBase/fp-go-loom/predicate/array"
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

type emailData struct {
	user     map[string]*user.User
	strains  []*template.StrainRows
	plasmids []*template.PlasmidRows
}

type mailgunEmailer struct {
	client    *mailgun.MailgunImpl
	logger    *logrus.Entry
	anno      *datasource.Annotation
	stk       *datasource.Stock
	usr       *datasource.User
	pub       publicationSource
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

func NewMailgunEmailer(args *EmailerParams) emailer.Handler {
	return &mailgunEmailer{
		name:      args.SenderName,
		from:      args.Sender,
		cc:        args.EmailCC,
		strprice:  args.StrainPrice,
		plasprice: args.PlasmidPrice,
		logger:    args.Logger,
		anno:      args.AnnoSource,
		stk:       args.StockSource,
		usr:       args.UserSource,
		pub:       args.PubSource,
		client:    getMailgunClient(args.Domain, args.APIKey),
	}
}

func (email *mailgunEmailer) SendEmail(ord *order.Order) error {
	all, body, err := email.emailBody(ord)
	if err != nil {
		email.logger.Error(err)
		return err
	}

	msg := email.client.NewMessage(
		fmt.Sprintf("%s <%s>", email.name, email.from),
		fmt.Sprintf(
			"Order ID:%s %s %s",
			ord.GetData().GetId(),
			all.user["shipper"].GetData().GetAttributes().GetFirstName(),
			all.user["shipper"].GetData().GetAttributes().GetLastName(),
		),
		fmt.Sprintf(
			etext,
			all.user["shipper"].GetData().GetAttributes().GetFirstName(),
			all.user["shipper"].GetData().GetAttributes().GetLastName(),
			ord.GetData().GetId(),
		),
	)

	err = msg.AddRecipient(all.user["shipper"].GetData().GetAttributes().GetEmail())
	if err != nil {
		email.logger.Error(err)
		return err
	}

	msg.AddCC(email.cc)

	if all.user["shipper"].GetData().
		GetAttributes().
		GetEmail() !=
		all.user["payer"].GetData().
			GetAttributes().
			GetEmail() {
		msg.AddCC(all.user["payer"].GetData().GetAttributes().GetEmail())
	}

	msg.AddBufferAttachment(
		fmt.Sprintf("invoice-%s.pdf", ord.GetData().GetId()),
		body.Bytes(),
	)

	id, err := email.postEmail(msg)
	if err != nil {
		return err
	}

	email.logger.Infof("message sent with id %s", id)

	return nil
}

func (email *mailgunEmailer) orderData(ord *order.Order) (*emailData, error) {
	all := &emailData{}

	strData, err := email.strains(ord)
	if err != nil {
		email.logger.Error(err)
		return all, err
	}

	plasData, err := email.plasmids(ord)
	if err != nil {
		return all, err
	}

	um, err := email.usr.UsersInOrder(ord)
	if err != nil {
		return all, err
	}

	all.strains = strData
	all.plasmids = plasData
	all.user = um

	return all, nil
}

func (email *mailgunEmailer) emailBody(
	ord *order.Order,
) (*emailData, *bytes.Buffer, error) {
	var b *bytes.Buffer

	all, err := email.orderData(ord)
	if err != nil {
		return all, b, err
	}

	body, err := template.OutputPDF(&template.OutputParams{
		Path: "/",
		File: "email.tmpl",
		Content: &template.EmailContent{
			StrainData:  all.strains,
			PlasmidData: all.plasmids,
			Content: &template.Content{
				Order:        ord,
				Shipper:      all.user["shipper"],
				Payer:        all.user["payer"],
				StrainPrice:  email.strprice,
				PlasmidPrice: email.plasprice,
			},
		},
	})

	return all, body, err
}

func (email *mailgunEmailer) postEmail(msg *mailgun.Message) (string, error) {
	_, id, err := email.client.Send(context.Background(), msg)
	if err != nil {
		email.logger.Errorf("error in sending email %s", err)
		return id, fmt.Errorf("error in sending email %s", err)
	}

	return id, nil
}

func (email *mailgunEmailer) plasmids(
	ord *order.Order,
) ([]*template.PlasmidRows, error) {
	var prows []*template.PlasmidRows

	plasmids, err := email.stk.GetPlasmids(
		email.stk.StocksFromItems(ord, "DBP"),
	)
	if err != nil {
		return prows, fmt.Errorf("error in getting plasmids %s", err)
	}

	plsinfo, err := email.stk.GetBasicPlasmidInfo(plasmids)
	if err != nil {
		return prows, fmt.Errorf("error in getting plasmid information %s", err)
	}

	prows, err = E.UnwrapError(ioeutils.ToEither(
		email.addPlasmidPub(plsinfo, plasmids),
	))
	if err != nil {
		return prows, fmt.Errorf(
			"error in adding publication to plasmids %s",
			err,
		)
	}

	return prows, nil
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

func (email *mailgunEmailer) strains(
	ord *order.Order,
) ([]*template.StrainRows, error) {
	var srows []*template.StrainRows

	strains, err := email.stk.GetStrains(email.stk.StocksFromItems(ord, "DBS"))
	if err != nil {
		return srows, fmt.Errorf("error in getting strains %s", err)
	}

	strInfo, err := email.anno.GetBasicStrainInfo(strains)
	if err != nil {
		return srows, fmt.Errorf("error in getting strain information %s", err)
	}

	srows, err = E.UnwrapError(ioeutils.ToEither(
		email.addStrainPub(strInfo, strains),
	))
	if err != nil {
		return srows, fmt.Errorf("error in adding pub to strain %s", err)
	}

	return srows, nil
}

func (email *mailgunEmailer) addStrainPub(
	strInfo [][]string,
	strains []*stock.Strain,
) IOE.IOEither[error, []*template.StrainRows] {
	builder := rowBuilder{Emailer: email, Info: strInfo}
	traverse := IOE.TraverseArrayWithIndexSeq(builder.resolveStrain)

	return F.Pipe1(strains, traverse)
}

func (email *mailgunEmailer) pubInfo(
	ids []string,
) IOE.IOEither[error, []*datasource.PubInfo] {
	return F.Pipe2(
		ids,
		normalizePublicationIDs,
		IOE.TraverseArraySeq(email.pub.ParsedInfo),
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
