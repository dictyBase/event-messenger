package github

import (
	"context"
	"fmt"

	"github.com/dictyBase/event-messenger/internal/datasource"
	issue "github.com/dictyBase/event-messenger/internal/issue-tracker"

	"github.com/dictyBase/event-messenger/internal/template"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/dictyBase/go-genproto/dictybaseapis/user"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/sirupsen/logrus"
)

type allData struct {
	*strainData
	*plasmidData
	user map[string]*user.User
}

type strainData struct {
	strains []*stock.Strain
	invs    [][]string
	info    [][]string
}

type plasmidData struct {
	plasmids []*stock.Plasmid
	invs     [][]string
}

// githubIssue includes all of the data needed to create an issue.
type githubIssue struct {
	token      string
	owner      string
	repository string
	client     *github.Client
	logger     *logrus.Entry
	anno       *datasource.Annotation
	stk        *datasource.Stock
	usr        *datasource.User
}

type IssueParams struct {
	Token       string
	Owner       string
	Repository  string
	Logger      *logrus.Entry
	AnnoSource  *datasource.Annotation
	StockSource *datasource.Stock
	UserSource  *datasource.User
}

type postIssueParams struct {
	title  string
	body   string
	labels []string
}

// NewIssueCreator acts as a constructor for Github issue creation
func NewIssueCreator(args *IssueParams) issue.IssueTracker {
	tc := oauth2.NewClient(
		context.Background(),
		oauth2.StaticTokenSource(&oauth2.Token{AccessToken: args.Token}),
	)
	return &githubIssue{
		client:     github.NewClient(tc),
		token:      args.Token,
		owner:      args.Owner,
		repository: args.Repository,
		logger:     args.Logger,
		anno:       args.AnnoSource,
		stk:        args.StockSource,
		usr:        args.UserSource,
	}
}

// CreateIssue creates a new GitHub issue based on order data.
func (gh *githubIssue) CreateIssue(ord *order.Order) error {
	all, err := gh.orderData(ord)
	if err != nil {
		gh.logger.Error(err)
		return err
	}
	b, err := template.OutputText("./../../assets", "issue.tmpl", getContent(all, ord))
	if err != nil {
		gh.logger.Error(err)
		return err
	}
	issue, err := gh.postIssue(&postIssueParams{
		labels: gh.labels(
			all.strainData.strains,
			all.plasmidData.plasmids,
		),
		body:  b.String(),
		title: fmt.Sprintf("Order ID:%s %s", ord.Data.Id, ord.Data.Attributes.Purchaser),
	})
	if err != nil {
		gh.logger.Errorf("error in posting issue to github %s", err)
		return fmt.Errorf("error in posting issue to github %s", err)
	}
	gh.logger.Infof("created a new issue with id %s", issue.GetHTMLURL())
	return nil
}

func (gh *githubIssue) orderData(ord *order.Order) (*allData, error) {
	all := &allData{}
	strData, err := gh.strains(ord)
	if err != nil {
		return all, err
	}
	plasData, err := gh.plasmids(ord)
	if err != nil {
		return all, err
	}
	um, err := gh.usr.UsersInOrder(ord)
	if err != nil {
		return all, err
	}
	all.user = um
	all.strainData = strData
	all.plasmidData = plasData
	return all, nil
}

func (gh *githubIssue) labels(strains []*stock.Strain, plasmids []*stock.Plasmid) []string {
	var labels []string
	if len(strains) > 0 {
		labels = append(labels, "Strain")
	}
	if len(plasmids) > 0 {
		labels = append(labels, "Plasmid")
	}
	return labels
}

func (gh *githubIssue) postIssue(args *postIssueParams) (*github.Issue, error) {
	input := &github.IssueRequest{
		Title:  &args.title,
		Body:   &args.body,
		Labels: &args.labels,
	}
	iss, _, err := gh.client.Issues.Create(
		context.Background(),
		gh.owner,
		gh.repository,
		input,
	)
	return iss, err
}

func (gh *githubIssue) strains(ord *order.Order) (*strainData, error) {
	sd := &strainData{}
	strains, err := gh.stk.GetStrains(gh.stk.StocksFromItems(ord, "DBS"))
	if err != nil {
		return sd, fmt.Errorf("error in getting strains %s", err)
	}
	strInvs, err := gh.anno.GetStrainInv(strains)
	if err != nil {
		return sd, fmt.Errorf("error in getting strain inventories %s", err)
	}
	strInfo, err := gh.anno.GetStrainInfo(strains)
	if err != nil {
		return sd, fmt.Errorf("error in getting strain information %s", err)
	}
	sd.strains = strains
	sd.invs = strInvs
	sd.info = strInfo
	return sd, nil
}

func (gh *githubIssue) plasmids(ord *order.Order) (*plasmidData, error) {
	pd := &plasmidData{}
	plasmids, err := gh.stk.GetPlasmids(gh.stk.StocksFromItems(ord, "DBP"))
	if err != nil {
		return pd, fmt.Errorf("error in getting plasmids %s", err)
	}
	plasInv, err := gh.anno.GetPlasmidInv(plasmids)
	if err != nil {
		return pd, fmt.Errorf("error in getting plasmid inventories %s", err)
	}
	pd.plasmids = plasmids
	pd.invs = plasInv
	return pd, nil
}

func getContent(all *allData, ord *order.Order) *template.IssueContent {
	return &template.IssueContent{
		Strains:    all.strainData.strains,
		Plasmids:   all.plasmidData.plasmids,
		StrainInv:  all.strainData.invs,
		PlasmidInv: all.plasmidData.invs,
		StrainInfo: all.strainData.info,
		Content: &template.Content{
			Shipper: all.user["shipper"],
			Payer:   all.user["payer"],
			Order:   ord,
		},
	}
}
