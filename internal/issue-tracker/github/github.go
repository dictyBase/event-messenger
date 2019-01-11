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

// CreateIssue creates a Github issue when a new stock order comes through
func CreateIssue(c *cli.Context) error {
	s, err := nats.NewSubscriber(c.String("nats-host"), c.String("nats-port"))
	if err != nil {
		log.Fatalf("cannot connect to nats for subscription %s\n", err)
	}
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

func issueCreator(c *cli.Context, ord *order.Order) error {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.String("gh-token")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	var labels []string
	var body string
	// Loop through items purchased.
	// Right now it lists one item ID per line.
	// It also adds labels based on whether item is strain or plasmid.
	for _, a := range ord.Data.Attributes.Items {
		body = fmt.Sprintf("Item: %s\n", a)
		if strings.Contains(a, "DBS") {
			labels = append(labels, "Strain Order")
		}
		if strings.Contains(a, "DBP") {
			labels = append(labels, "Plasmid Order")
		}
	}
	// Generate Github issue title
	title := fmt.Sprintf("Order ID:%s %s", ord.Data.Id, ord.Data.Attributes.Purchaser)

	input := &github.IssueRequest{
		Title:  &title,
		Body:   &body,
		Labels: &labels,
	}

	_, _, err := client.Issues.Create(ctx, c.String("owner"), c.String("repository"), input)
	if err != nil {
		return fmt.Errorf("error in creating github issue %s", err)
	}
	return nil
}
