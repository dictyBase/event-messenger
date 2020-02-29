package github

import (
	"context"
	"fmt"
	"strings"

	issue "github.com/dictyBase/event-messenger/internal/issue-tracker"

	"github.com/dictyBase/go-genproto/dictybaseapis/annotation"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
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
}

type IssueParams struct {
	Token       string
	Owner       string
	Repository  string
	Logger      *logrus.Entry
	AnnoClient  annotation.TaggedAnnotationServiceClient
	StockClient stock.StockServiceClient
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
	}
}

// CreateIssue creates a new GitHub issue based on order data.
func (gh *githubIssue) CreateIssue(ord *order.Order) error {
	var labels []string
	var str strings.Builder
	for _, a := range ord.Data.Attributes.Items {
		str.WriteString("Item: ")
		str.WriteString(a)
		str.WriteString("\n")
		if strings.Contains(a, "DBS") {
			labels = append(labels, "Strain Order")
		}
		if strings.Contains(a, "DBP") {
			labels = append(labels, "Plasmid Order")
		}
	}
	title := fmt.Sprintf("Order ID:%s %s", ord.Data.Id, ord.Data.Attributes.Purchaser)
	body := str.String()
	input := &github.IssueRequest{
		Title:  &title,
		Body:   &body,
		Labels: &labels,
	}
	iss, _, err := gh.client.Issues.Create(context.Background(), gh.owner, gh.repository, input)
	if err != nil {
		gh.logger.Errorf("error in creating github issue %s", err)
		return fmt.Errorf("error in creating github issue %s", err)
	}
	gh.logger.Infof("created a new issue with id %s", *iss.URL)
	return nil
}
