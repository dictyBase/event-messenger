package nats

import (
	"fmt"

	"github.com/dictyBase/event-messenger/internal/message"
	email "github.com/dictyBase/event-messenger/internal/send-email"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	gnats "github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats/encoders/protobuf"
	"github.com/sirupsen/logrus"
)

type EmailSubscriber struct {
	econn  *gnats.EncodedConn
	logger *logrus.Entry
}

// NewEmailSubscriber connects to nats
func NewEmailSubscriber(
	host, port string,
	logger *logrus.Entry,
	options ...gnats.Option,
) (*EmailSubscriber, error) {
	nc, err := gnats.Connect(
		fmt.Sprintf("nats://%s:%s", host, port),
		options...)
	if err != nil {
		return &EmailSubscriber{}, err
	}

	ec, err := gnats.NewEncodedConn(nc, protobuf.PROTOBUF_ENCODER)
	if err != nil {
		return &EmailSubscriber{}, err
	}

	return &EmailSubscriber{econn: ec, logger: logger}, nil
}

// Start starts the subscription server and handles the incoming stock order data.
func (n *EmailSubscriber) Start(
	sub string,
	client email.Handler,
) error {
	_, err := n.econn.Subscribe(sub, func(ord *order.Order) {
		if err := client.SendEmail(ord); err != nil {
			n.logger.Error(err)
		}
	})

	return message.HandleConnection(n.econn, err)
}

// Stop stops the server
func (n *EmailSubscriber) Stop() error {
	n.econn.Close()
	return nil
}
