package mailgun

import (
	"fmt"

	"github.com/dictyBase/event-messenger/internal/datasource"
	emailer "github.com/dictyBase/event-messenger/internal/send-email"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/mailgun/mailgun-go/v3"
	"github.com/sirupsen/logrus"
)

type mailgunEmailer struct {
	client *mailgun.MailgunImpl
	logger *logrus.Entry
	anno   *datasource.Annotation
	stk    *datasource.Stock
	usr    *datasource.User
	pub    *datasource.Publication
}

type EmailerParams struct {
	Domain      string
	ApiKey      string
	Loger       *logrus.Entry
	AnnoSource  *datasource.Annotation
	StockSource *datasource.Stock
	UserSource  *datasource.User
	PubSource   *datasource.Publication
}

type strainRow struct {
	Id         string
	SysName    string
	Names      string
	Descriptor string
	PubInfo    []*datasource.PubInfo
}

type plasmidRow struct {
	Id      string
	Name    string
	PubInfo []*datasource.PubInfo
}

func NewMailgunEmailer(args *EmailerParams) emailer.EmailHandler {
	return &mailgunEmailer{
		logger: args.Loger,
		anno:   args.AnnoSource,
		stk:    args.StockSource,
		usr:    args.UserSource,
		pub:    args.PubSource,
		client: getMailgunClient(args.Domain, args.ApiKey),
	}
}

func (email *mailgunEmailer) SendEmail(ord *order.Order) error {
	return nil

}

func (email *mailgunEmailer) plasmids(ord *order.Order) ([]*plasmidRow, error) {
	var prows []*plasmidRow
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

func (email *mailgunEmailer) addPlasmidPub(strInfo [][]string, plasmids []*stock.Plasmid) ([]*plasmidRow, error) {
	var prows []*plasmidRow
	for i, pls := range plasmids {
		pinfo, err := email.pubInfo(pls.Data.Attributes.Publications)
		if err != nil {
			return prows, err
		}
		prows = append(prows, &plasmidRow{
			Id:      strInfo[i][0],
			Name:    strInfo[i][2],
			PubInfo: pinfo,
		})
	}
	return prows, nil
}

func (email *mailgunEmailer) strains(ord *order.Order) ([]*strainRow, error) {
	var srows []*strainRow
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

func (email *mailgunEmailer) addStrainPub(strInfo [][]string, strains []*stock.Strain) ([]*strainRow, error) {
	var srows []*strainRow
	for i, str := range strains {
		pinfo, err := email.pubInfo(str.Data.Attributes.Publications)
		if err != nil {
			return srows, err
		}
		srows = append(srows, &strainRow{
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
