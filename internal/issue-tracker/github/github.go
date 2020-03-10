package github

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"math/rand"
	"time"

	issue "github.com/dictyBase/event-messenger/internal/issue-tracker"

	"github.com/dictyBase/go-genproto/dictybaseapis/annotation"
	"github.com/dictyBase/go-genproto/dictybaseapis/api/jsonapi"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/dictyBase/go-genproto/dictybaseapis/user"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/sirupsen/logrus"
)

// githubIssue includes all of the data needed to create an issue.
type githubIssue struct {
	token      string
	owner      string
	repository string
	client     *github.Client
	logger     *logrus.Entry
	aclient    annotation.TaggedAnnotationServiceClient
	sclient    stock.StockServiceClient
	uclient    user.UserServiceClient
}

type IssueParams struct {
	Token       string
	Owner       string
	Repository  string
	Logger      *logrus.Entry
	AnnoClient  annotation.TaggedAnnotationServiceClient
	StockClient stock.StockServiceClient
	UserClient  user.UserServiceClient
}

type postIssueParams struct {
	title  string
	body   string
	owner  string
	repo   string
	labels []string
	client *github.Client
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
		aclient:    args.AnnoClient,
		sclient:    args.StockClient,
		uclient:    args.UserClient,
	}
}

func (gh *githubIssue) usersInOrder(ord *order.Order) (map[string]*user.User, error) {
	m := make(map[string]*user.User)
	pu, err := gh.uclient.GetUserByEmail(
		context.Background(),
		&jsonapi.GetEmailRequest{Email: ord.Data.Attributes.Payer},
	)
	if err != nil {
		return m, err
	}
	su, err := gh.uclient.GetUserByEmail(
		context.Background(),
		&jsonapi.GetEmailRequest{Email: ord.Data.Attributes.Consumer},
	)
	if err != nil {
		return m, err
	}
	m["payer"] = pu
	m["shipper"] = su
	return m, nil
}

// CreateIssue creates a new GitHub issue based on order data.
func (gh *githubIssue) CreateIssue(ord *order.Order) error {
	t, err := template.New("stock-invoice").Parse(tmpl)
	if err != nil {
		gh.logger.Errorf("error in parsing template %s", err)
		return fmt.Errorf("error in parsing template %s", err)
		return err
	}
	strains, err := getStrains(strainsFromItems(ord), gh.sclient)
	if err != nil {
		gh.logger.Errorf("error in getting strains %s", err)
		return fmt.Errorf("error in getting strains %s", err)
	}
	strInvs, err := getStrainInv(strains, gh.aclient)
	if err != nil {
		gh.logger.Errorf("error in getting strain inventories %s", err)
		return fmt.Errorf("error in getting strain inventories %s", err)
	}
	strInfo, err := getStrainInfo(strains, gh.aclient)
	if err != nil {
		gh.logger.Errorf("error in getting strain information %s", err)
		return fmt.Errorf("error in getting strain information %s", err)
	}
	plasmids, err := getPlasmids(plasmidsFromItems(ord), gh.sclient)
	if err != nil {
		gh.logger.Errorf("error in getting plasmids %s", err)
		return fmt.Errorf("error in getting plasmids %s", err)
	}
	plasInv, err := getPlasmidInv(plasmids, gh.aclient)
	if err != nil {
		gh.logger.Errorf("error in getting plasmid inventories %s", err)
		return fmt.Errorf("error in getting plasmid inventories %s", err)
	}
	cont := &IssueContent{
		Strains:    strains,
		Plasmids:   plasmids,
		StrainInv:  strInvs,
		PlasmidInv: plasInv,
		StrainInfo: strInfo,
		Order:      ord,
	}
	var b bytes.Buffer
	if err := t.Execute(&b, cont); err != nil {
		gh.logger.Errorf("error in executing template %s", err)
		return fmt.Errorf("error in executing template %s", err)
	}
	var labels []string
	if len(strains) > 0 {
		labels = append(labels, "Strain")
	}
	if len(plasmids) > 0 {
		labels = append(labels, "Plasmid")
	}
	issue, err := postIssue(&postIssueParams{
		client: gh.client,
		labels: labels,
		body:   b.String(),
		title:  fmt.Sprintf("Order ID:%s %s", ord.Data.Id, ord.Data.Attributes.Purchaser),
		owner:  gh.owner,
		repo:   gh.repository,
	})

	gh.logger.Infof("created a new issue with id %s", *iss.URL)
	return nil
}

func postIssue(args *postIssueParams) (*github.Issue, error) {
	input := &github.IssueRequest{
		Title:  &args.title,
		Body:   &args.body,
		Labels: &args.labels,
	}
	iss, _, err := args.client.Issues.Create(
		context.Background(),
		args.owner,
		args.repo,
		input,
	)
	return iss, err
}

func randNum(min, max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return min + r.Intn(max-min)
}
