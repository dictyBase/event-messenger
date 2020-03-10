package github

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"math/rand"
	"time"

	"github.com/dictyBase/event-messenger/internal/datasource"
	issue "github.com/dictyBase/event-messenger/internal/issue-tracker"

	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/sirupsen/logrus"
)

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
	t, err := template.New("stock-invoice").Parse(tmpl)
	if err != nil {
		gh.logger.Errorf("error in parsing template %s", err)
		return fmt.Errorf("error in parsing template %s", err)
	}
	strData, err := gh.strains(ord)
	if err != nil {
		gh.logger.Error(err.Error())
		return err
	}
	plasData, err := gh.plasmids(ord)
	if err != nil {
		gh.logger.Error(err.Error())
		return err
	}
	um, err := gh.usr.UsersInOrder(ord)
	if err != nil {
		gh.logger.Error(err.Error())
		return err
	}
	cont := &IssueContent{
		Strains:    strData.strains,
		Plasmids:   plasData.plasmids,
		StrainInv:  strData.invs,
		PlasmidInv: plasData.invs,
		StrainInfo: strData.info,
		Shipper:    um["shipper"],
		Payer:      um["payer"],
		Order:      ord,
	}
	var b bytes.Buffer
	if err := t.Execute(&b, cont); err != nil {
		gh.logger.Errorf("error in executing template %s", err)
		return fmt.Errorf("error in executing template %s", err)
	}
	issue, err := gh.postIssue(&postIssueParams{
		labels: gh.labels(strData.strains, plasData.plasmids),
		body:   b.String(),
		title:  fmt.Sprintf("Order ID:%s %s", ord.Data.Id, ord.Data.Attributes.Purchaser),
	})
	gh.logger.Infof("created a new issue with id %s", issue.GetHTMLURL())
	return nil
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
	strains, err := gh.stk.GetStrains(gh.stk.StrainsFromItems(ord))
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
	plasmids, err := gh.stk.GetPlasmids(gh.stk.PlasmidsFromItems(ord))
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

func randNum(min, max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return min + r.Intn(max-min)
}
