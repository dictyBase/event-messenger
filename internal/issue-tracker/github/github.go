package github

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/dictyBase/event-messenger/internal/message"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/Sirupsen/logrus"

	"github.com/dictyBase/event-messenger/internal/message/nats"

	cli "gopkg.in/urfave/cli.v1"
)

/**
TODO:
1) Need to return issue.IssueTracker in NewIssueCreator
2) Figure out how to properly pass CreateIssue into Start
3) Need to write CreateIssue, maybe by merging issueContentCreator?
*/

// Issue includes all of the data needed to create an issue.
type Issue struct {
	token      string
	owner      string
	repository string
}

// NewIssueCreator acts as a constructor for Github issue creation
func NewIssueCreator(s *Issue) error {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: s.token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	n, err := issueContentCreator(ord)
	if err != nil {
		return fmt.Errorf("error in creating github issue format %s", err)
	}

	_, _, err := client.Issues.Create(ctx, s.owner, s.repository, n)
	if err != nil {
		return fmt.Errorf("error in creating github issue %s", err)
	}
	return nil
}

// func (iss *Issue) CreateIssue(ord *order.Order) error {

// }

// RunCreateIssue creates a Github issue when a new stock order comes through
func RunCreateIssue(c *cli.Context) error {
	s, err := nats.NewSubscriber(c.String("nats-host"), c.String("nats-port"))
	if err != nil {
		log.Fatalf("cannot connect to nats for subscription %s\n", err)
	}
	iss := &Issue{token: c.String("gh-token"), owner: c.String("owner"), repository: c.String("repository")}
	// g := NewIssueCreator(iss)
	// if g != nil {
	// 	return fmt.Errorf("error in getting github flags %s", err)
	// }
	// err = s.Start(
	// 	"OrderService.*",
	// 	issue.IssueTracker,
	// )
	// if err != nil {
	// 	return cli.NewExitError(
	// 		fmt.Sprintf("cannot start the subscriber server %s", err),
	// 		2,
	// 	)
	// }
	logger := getLogger(c)
	logger.Info("starting the Github issue creation subscriber messaging backend")
	shutdown(s, logger)
	return nil
}

// issueContentCreator generates the content to post in the Github issue
func issueContentCreator(ord *order.Order) (*github.IssueRequest, error) {
	var labels []string
	var body string
	for _, a := range ord.Data.Attributes.Items {
		body = fmt.Sprintf("Item: %s\n", a)
		if strings.Contains(a, "DBS") {
			labels = append(labels, "Strain Order")
		}
		if strings.Contains(a, "DBP") {
			labels = append(labels, "Plasmid Order")
		}
	}
	title := fmt.Sprintf("Order ID:%s %s", ord.Data.Id, ord.Data.Attributes.Purchaser)
	input := &github.IssueRequest{
		Title:  &title,
		Body:   &body,
		Labels: &labels,
	}
	return input, nil
}

func getLogger(c *cli.Context) *logrus.Entry {
	log := logrus.New()
	log.Out = os.Stderr
	switch c.GlobalString("log-format") {
	case "text":
		log.Formatter = &logrus.TextFormatter{
			TimestampFormat: "02/Jan/2006:15:04:05",
		}
	case "json":
		log.Formatter = &logrus.JSONFormatter{
			TimestampFormat: "02/Jan/2006:15:04:05",
		}
	}
	l := c.GlobalString("log-level")
	switch l {
	case "debug":
		log.Level = logrus.DebugLevel
	case "warn":
		log.Level = logrus.WarnLevel
	case "error":
		log.Level = logrus.ErrorLevel
	case "fatal":
		log.Level = logrus.FatalLevel
	case "panic":
		log.Level = logrus.PanicLevel
	}
	return logrus.NewEntry(log)
}

func shutdown(r message.Subscriber, logger *logrus.Entry) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	<-ch
	logger.Info("received kill signal")
	if err := r.Stop(); err != nil {
		logger.Fatalf("unable to close the subscription %s\n", err)
	}
	logger.Info("closed the connections gracefully")
}
