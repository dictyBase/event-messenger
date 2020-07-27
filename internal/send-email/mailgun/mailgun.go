package mailgun

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/dictyBase/event-messenger/internal/datasource"
	emailer "github.com/dictyBase/event-messenger/internal/send-email"
	"github.com/dictyBase/event-messenger/internal/template"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/dictyBase/go-genproto/dictybaseapis/user"
	"github.com/mailgun/mailgun-go/v3"
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
	APIKey       string
	StrainPrice  int
	PlasmidPrice int
	Logger       *logrus.Entry
	PubSource    *datasource.Publication
	*datasource.Sources
}

func NewMailgunEmailer(args *EmailerParams) emailer.EmailHandler {
	return &mailgunEmailer{
		name:      args.SenderName,
		from:      args.Sender,
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

func (email *mailgunEmailer) emailBody(ord *order.Order) (*emailData, *bytes.Buffer, error) {
	var b *bytes.Buffer
	all, err := email.orderData(ord)
	if err != nil {
		return all, b, err
	}
	body, err := template.OutputHTML(&template.OutputParams{
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
		}})
	return all, body, err
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
			ord.Data.Id,
			all.user["shipper"].Data.Attributes.FirstName,
			all.user["shipper"].Data.Attributes.LastName,
		),
		etext,
	)
	err = msg.AddRecipient(all.user["shipper"].Data.Attributes.Email)
	if err != nil {
		email.logger.Error(err)
		return err
	}
	if all.user["shipper"].Data.Attributes.Email != all.user["payer"].Data.Attributes.Email {
		msg.AddCC(all.user["payer"].Data.Attributes.Email)
	}
	msg.SetHtml(body.String())
	id, err := email.postEmail(msg)
	if err != nil {
		return err
	}
	email.logger.Infof("message send with id %s", id)
	return nil
}

func (email *mailgunEmailer) postEmail(msg *mailgun.Message) (string, error) {
	_, id, err := email.client.Send(context.Background(), msg)
	if err != nil {
		email.logger.Errorf("error in sending email %s", err)
		return id, fmt.Errorf("error in sending email %s", err)
	}
	return id, nil
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
		prows = append(prows, &template.PlasmidRows{
			ID:   strInfo[i][0],
			Name: strInfo[i][2],
		})
		if len(pls.Data.Attributes.Publications) == 0 {
			continue
		}
		pinfo, err := email.pubInfo(pls.Data.Attributes.Publications)
		if err != nil {
			return prows, err
		}
		prows[i].PubInfo = pinfo
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
		srows = append(srows, &template.StrainRows{
			ID:         strInfo[i][0],
			Descriptor: strInfo[i][1],
			Names:      strInfo[i][2],
			SysName:    strInfo[i][3],
		})
		if len(str.Data.Attributes.Publications) == 0 {
			continue
		}
		pinfo, err := email.pubInfo(str.Data.Attributes.Publications)
		if err != nil {
			return srows, err
		}
		srows[i].PubInfo = pinfo
	}
	return srows, nil
}

func (email *mailgunEmailer) pubInfo(ids []string) ([]*datasource.PubInfo, error) {
	var pinfo []*datasource.PubInfo
	for _, pid := range ids {
		if len(strings.TrimSpace(pid)) == 0 {
			continue
		}
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
