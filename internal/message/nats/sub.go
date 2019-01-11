package nats

import (
	"context"
	"fmt"
	"strings"

	"github.com/dictyBase/event-messenger/internal/message"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/dictyBase/go-genproto/dictybaseapis/order"

	gnats "github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats/encoders/protobuf"
)

type natsSubscriber struct {
	econn *gnats.EncodedConn
}

// NewSubscriber connects to nats
func NewSubscriber(host, port string, options ...gnats.Option) (message.NatsSubscriber, error) {
	nc, err := gnats.Connect(fmt.Sprintf("nats://%s:%s", host, port), options...)
	if err != nil {
		return &natsSubscriber{}, err
	}
	ec, err := gnats.NewEncodedConn(nc, protobuf.PROTOBUF_ENCODER)
	if err != nil {
		return &natsSubscriber{}, err
	}
	return &natsSubscriber{econn: ec}, nil
}

// Start starts the server and communicates using a channel
func (n *natsSubscriber) Start(sub string, client message.IssueTracker) error {
	_, err := n.econn.Subscribe(sub, func(c *cli.Context, ord *order.Order) {
		issueCreator(c, ord)
	})
	if err != nil {
		return err
	}
	if err := n.econn.Flush(); err != nil {
		return err
	}
	if err := n.econn.LastError(); err != nil {
		return err
	}
	return nil
}

// Stop stops the server
func (n *natsSubscriber) Stop() error {
	n.econn.Close()
	return nil
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
