package nats

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	issue "github.com/dictyBase/event-messenger/internal/issue-tracker"
	"github.com/dictyBase/event-messenger/internal/message"

	"github.com/dictyBase/go-genproto/dictybaseapis/order"

	gnats "github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats/encoders/protobuf"
)

type natsSubscriber struct {
	econn  *gnats.EncodedConn
	logger *logrus.Logger
}

// NewSubscriber connects to nats
func NewSubscriber(host, port string, options ...gnats.Option) (message.Subscriber, error) {
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
func (n *natsSubscriber) Start(sub string, client issue.IssueTracker) error {
	_, err := n.econn.Subscribe(sub, func(ord *order.Order) {
		client.CreateIssue(ord)
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
