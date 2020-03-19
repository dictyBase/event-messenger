package mailgun

import (
	"bytes"
	"context"
	"fmt"
	html "html/template"
	"io/ioutil"

	"github.com/dictyBase/event-messenger/internal/datasource"
	emailer "github.com/dictyBase/event-messenger/internal/send-email"
	"github.com/dictyBase/event-messenger/internal/template"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/dictyBase/go-genproto/dictybaseapis/user"
	"github.com/mailgun/mailgun-go/v3"
	"github.com/markbates/pkger"
	"github.com/sirupsen/logrus"
)

const (
	etext = `In case no order information is available
			 below read the pdf attachment
			 `
)

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
	pub       *datasource.Publication
	strprice  int
	plasprice int
	from      string
	name      string
}

type EmailerParams struct {
	Sender       string
	SenderName   string
	Domain       string
	ApiKey       string
	StrainPrice  int
	PlasmidPrice int
	Loger        *logrus.Entry
	AnnoSource   *datasource.Annotation
	StockSource  *datasource.Stock
	UserSource   *datasource.User
	PubSource    *datasource.Publication
}

func NewMailgunEmailer(args *EmailerParams) emailer.EmailHandler {
	return &mailgunEmailer{
		name:      args.SenderName,
		from:      args.Sender,
		strprice:  args.StrainPrice,
		plasprice: args.PlasmidPrice,
		logger:    args.Loger,
		anno:      args.AnnoSource,
		stk:       args.StockSource,
		usr:       args.UserSource,
		pub:       args.PubSource,
		client:    getMailgunClient(args.Domain, args.ApiKey),
	}
}

func (email *mailgunEmailer) SendEmail(ord *order.Order) error {
	strData, err := email.strains(ord)
	if err != nil {
		email.logger.Error(err)
		return err
	}
	plasData, err := email.plasmids(ord)
	if err != nil {
		return err
	}
	um, err := email.usr.UsersInOrder(ord)
	if err != nil {
		return err
	}
	body, err := email.runTemplate(&template.EmailContent{
		StrainData:  strData,
		PlasmidData: plasData,
		Content: &template.Content{
			Order:        ord,
			Shipper:      um["shipper"],
			Payer:        um["payer"],
			StrainPrice:  email.strprice,
			PlasmidPrice: email.plasprice,
		},
	})
	if err != nil {
		return err
	}
	msg := email.client.NewMessage(
		fmt.Sprintf("%s<%s>", email.name, email.from),
		fmt.Sprintf(
			"Order ID:%s %s %s",
			ord.Data.Id,
			um["shipper"].Data.Attributes.FirstName,
			um["shipper"].Data.Attributes.LastName,
		),
		etext,
	)
	msg.AddRecipient(um["shipper"].Data.Attributes.Email)
	if um["shipper"].Data.Attributes.Email != um["payer"].Data.Attributes.Email {
		msg.AddCC(um["payer"].Data.Attributes.Email)
	}
	msg.SetHtml(body.String())
	_, id, err := email.client.Send(context.Background(), msg)
	if err != nil {
		email.logger.Errorf("error in sending email %s", err)
		return fmt.Errorf("error in sending email %s", err)
	}
	email.logger.Infof("message send with id %s", id)
	return nil
}

func (email *mailgunEmailer) runTemplate(cont *template.EmailContent) (bytes.Buffer, error) {
	var b bytes.Buffer
	tb, err := email.readTemplate("/assets/email.tmpl")
	if err != nil {
		return b, err
	}
	t, err := html.New("stock-invoice").Parse(string(tb))
	if err != nil {
		return b, fmt.Errorf("error in parsing template %s", err)
	}
	if err := t.Execute(&b, cont); err != nil {
		return b, fmt.Errorf("error in executing template %s", err)
	}
	return b, nil
}

func (email *mailgunEmailer) readTemplate(path string) ([]byte, error) {
	var b []byte
	f, err := pkger.Open(path)
	if err != nil {
		return b, fmt.Errorf("error in template file %s", err)
	}
	defer f.Close()
	tb, err := ioutil.ReadAll(f)
	if err != nil {
		return b, fmt.Errorf("error in reading template file content %s", err)
	}
	return tb, nil
}

func (email *mailgunEmailer) plasmids(ord *order.Order) ([]*template.PlasmidRows, error) {
	var prows []*template.PlasmidRows
	plasmids, err := email.stk.GetPlasmids(email.stk.StocksFromItems(ord, "DBP"))
	if err != nil {
		return prows, fmt.Errorf("error in getting plasmids %s", err)
	}
	plsinfo, err := email.stk.GetBasicPlasmidInfo(plasmids)
	if err != nil {
		return prows, fmt.Errorf("error in getting plasmid information %s", err)
	}
	prows, err = email.addPlasmidPub(plsinfo, plasmids)
	if err != nil {
		return prows, fmt.Errorf("error in adding publication to plasmids %s", err)
	}
	return prows, nil
}

func (email *mailgunEmailer) addPlasmidPub(strInfo [][]string, plasmids []*stock.Plasmid) ([]*template.PlasmidRows, error) {
	var prows []*template.PlasmidRows
	for i, pls := range plasmids {
		pinfo, err := email.pubInfo(pls.Data.Attributes.Publications)
		if err != nil {
			return prows, err
		}
		prows = append(prows, &template.PlasmidRows{
			Id:      strInfo[i][0],
			Name:    strInfo[i][2],
			PubInfo: pinfo,
		})
	}
	return prows, nil
}

func (email *mailgunEmailer) strains(ord *order.Order) ([]*template.StrainRows, error) {
	var srows []*template.StrainRows
	strains, err := email.stk.GetStrains(email.stk.StocksFromItems(ord, "DBS"))
	if err != nil {
		return srows, fmt.Errorf("error in getting strains %s", err)
	}
	strInfo, err := email.anno.GetBasicStrainInfo(strains)
	if err != nil {
		return srows, fmt.Errorf("error in getting strain information %s", err)
	}
	srows, err = email.addStrainPub(strInfo, strains)
	if err != nil {
		return srows, fmt.Errorf("error in adding pub to strain %s", err)
	}
	return srows, nil
}

func (email *mailgunEmailer) addStrainPub(strInfo [][]string, strains []*stock.Strain) ([]*template.StrainRows, error) {
	var srows []*template.StrainRows
	for i, str := range strains {
		pinfo, err := email.pubInfo(str.Data.Attributes.Publications)
		if err != nil {
			return srows, err
		}
		srows = append(srows, &template.StrainRows{
			Id:         strInfo[i][0],
			Descriptor: strInfo[i][1],
			Names:      strInfo[i][2],
			SysName:    strInfo[i][3],
			PubInfo:    pinfo,
		})
	}
	return srows, nil
}

func (email *mailgunEmailer) pubInfo(ids []string) ([]*datasource.PubInfo, error) {
	var pinfo []*datasource.PubInfo
	for _, pid := range ids {
		pub, err := email.pub.ParsedInfo(pid)
		if err != nil {
			return pinfo, err
		}
		pinfo = append(pinfo, pub)
	}
	return pinfo, nil
}

func getMailgunClient(domain, apiKey string) *mailgun.MailgunImpl {
	return mailgun.NewMailgun(domain, apiKey)
}
