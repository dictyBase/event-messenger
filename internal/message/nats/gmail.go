package nats

import (
	"fmt"

	email "github.com/dictyBase/event-messenger/internal/send-email"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	gnats "github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats/encoders/protobuf"
	"github.com/sirupsen/logrus"
)

type NatsGmailSubscriber struct {
	econn  *gnats.EncodedConn
	logger *logrus.Entry
}

// NewGmailSubscriber connects to nats
func NewGmailSubscriber(host, port string, logger *logrus.Entry, options ...gnats.Option) (*NatsGmailSubscriber, error) {
	nc, err := gnats.Connect(fmt.Sprintf("nats://%s:%s", host, port), options...)
	if err != nil {
		return &NatsGmailSubscriber{}, err
	}
	ec, err := gnats.NewEncodedConn(nc, protobuf.PROTOBUF_ENCODER)
	if err != nil {
		return &NatsGmailSubscriber{}, err
	}
	return &NatsGmailSubscriber{econn: ec, logger: logger}, nil
}

// Start starts the subscription server and handles the incoming stock order data.
func (n *NatsGmailSubscriber) Start(sub string, client email.EmailHandler) error {
	_, err := n.econn.Subscribe(sub, func(ord *order.Order) {
		if err := client.SendEmail(ord); err != nil {
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
func (n *NatsGmailSubscriber) Stop() error {
	n.econn.Close()
	return nil
}
