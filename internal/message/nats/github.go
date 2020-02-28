package nats

import (
	"fmt"

	issue "github.com/dictyBase/event-messenger/internal/issue-tracker"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	gnats "github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats/encoders/protobuf"
	"github.com/sirupsen/logrus"
)

type NatsGithubSubscriber struct {
	econn  *gnats.EncodedConn
	logger *logrus.Entry
}

// NewGithubSubscriber connects to nats
func NewGithubSubscriber(host, port string, logger *logrus.Entry, options ...gnats.Option) (*NatsGithubSubscriber, error) {
	nc, err := gnats.Connect(fmt.Sprintf("nats://%s:%s", host, port), options...)
	if err != nil {
		return &NatsGithubSubscriber{}, err
	}
	ec, err := gnats.NewEncodedConn(nc, protobuf.PROTOBUF_ENCODER)
	if err != nil {
		return &NatsGithubSubscriber{}, err
	}
	return &NatsGithubSubscriber{econn: ec, logger: logger}, nil
}

// Start starts the subscription server and handles the incoming stock order data.
func (n *NatsGithubSubscriber) Start(sub string, client issue.IssueTracker) error {
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
func (n *NatsGithubSubscriber) Stop() error {
	n.econn.Close()
	return nil
}
