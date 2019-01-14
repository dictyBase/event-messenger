package github

import (
	"context"
	"fmt"
	"strings"

	issue "github.com/dictyBase/event-messenger/internal/issue-tracker"

	"github.com/dictyBase/go-genproto/dictybaseapis/order"
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
}

// NewIssueCreator acts as a constructor for Github issue creation
func NewIssueCreator(token, owner, repository string, logger *logrus.Entry) issue.IssueTracker {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return &githubIssue{token: token, owner: owner, repository: repository, client: client, logger: logger}
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
