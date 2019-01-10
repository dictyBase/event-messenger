package nats

import (
	"fmt"

	"github.com/dictyBase/event-messenger/internal/message"

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
	_, err := n.econn.Subscribe(sub, func(sub string, ord *order.Order) {
		// this is where the issue needs to be created
		// need get the obj from protocol buffer and then use that data
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
