package nats

import (
	"fmt"

	issue "github.com/dictyBase/event-messenger/internal/issue-tracker"
	"github.com/dictyBase/event-messenger/internal/message"
	"github.com/sirupsen/logrus"

	"github.com/dictyBase/go-genproto/dictybaseapis/order"

	gnats "github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats/encoders/protobuf"
)

type natsGithubSubscriber struct {
	econn  *gnats.EncodedConn
	logger *logrus.Entry
}

// NewGithubSubscriber connects to nats
func NewGithubSubscriber(host, port string, logger *logrus.Entry, options ...gnats.Option) (message.GithubSubscriber, error) {
	nc, err := gnats.Connect(fmt.Sprintf("nats://%s:%s", host, port), options...)
	if err != nil {
		return &natsGithubSubscriber{}, err
	}
	ec, err := gnats.NewEncodedConn(nc, protobuf.PROTOBUF_ENCODER)
	if err != nil {
		return &natsGithubSubscriber{}, err
	}
	return &natsGithubSubscriber{econn: ec, logger: logger}, nil
}

// Start starts the subscription server and handles the incoming stock order data.
func (n *natsGithubSubscriber) Start(sub string, client issue.IssueTracker) error {
	_, err := n.econn.Subscribe(sub, func(ord *order.Order) {
		if err := client.CreateIssue(ord); err != nil {
			n.logger.Error(err)
		}
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
func (n *natsGithubSubscriber) Stop() error {
	n.econn.Close()
	return nil
}
