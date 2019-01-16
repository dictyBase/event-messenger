package nats

import (
	"fmt"

	email "github.com/dictyBase/event-messenger/internal/send-email"

	"github.com/dictyBase/event-messenger/internal/message"
	"github.com/sirupsen/logrus"

	"github.com/dictyBase/go-genproto/dictybaseapis/order"

	gnats "github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats/encoders/protobuf"
)

type natsGmailSubscriber struct {
	econn  *gnats.EncodedConn
	logger *logrus.Entry
}

// NewGmailSubscriber connects to nats
func NewGmailSubscriber(host, port string, logger *logrus.Entry, options ...gnats.Option) (message.GmailSubscriber, error) {
	nc, err := gnats.Connect(fmt.Sprintf("nats://%s:%s", host, port), options...)
	if err != nil {
		return &natsGmailSubscriber{}, err
	}
	ec, err := gnats.NewEncodedConn(nc, protobuf.PROTOBUF_ENCODER)
	if err != nil {
		return &natsGmailSubscriber{}, err
	}
	return &natsGmailSubscriber{econn: ec, logger: logger}, nil
}

// Start starts the subscription server and handles the incoming stock order data.
func (n *natsGmailSubscriber) Start(sub string, client email.EmailHandler) error {
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
func (n *natsGmailSubscriber) Stop() error {
	n.econn.Close()
	return nil
}
