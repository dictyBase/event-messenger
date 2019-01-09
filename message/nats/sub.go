package nats

import (
	"github.com/dictyBase/event-messenger/message"

	"github.com/dictyBase/go-genproto/dictybaseapis/order"

	gnats "github.com/nats-io/go-nats"
)

type natsMessage struct {
	msg *gnats.Msg
}

func (m *natsMessage) Message() []byte {
	return m.msg.Data
}

func (m *natsMessage) Done() error {
	if err := m.msg.Sub.Unsubscribe(); err != nil {
		return err
	}
	return nil
}

type natsSubscriber struct {
	conn   *gnats.Conn
	err    error
	output chan message.SubscriberMessage
}

// NewSubscriber connects to nats
func NewSubscriber(url string, options ...gnats.Option) (message.Subscriber, error) {
	nc, err := gnats.Connect(url, options...)
	if err != nil {
		return &natsSubscriber{}, err
	}
	return &natsSubscriber{
		conn:   nc,
		output: make(chan message.SubscriberMessage),
	}, err
}

// Start starts the server and communicates using a channel
func (s *natsSubscriber) Start(sub string, ord *order.Order) <-chan message.SubscriberMessage {
	nm := &natsMessage{}
	_, err := s.conn.Subscribe(sub, func(msg *gnats.Msg) {
		// this is where the issue needs to be created
		// need get the obj from protocol buffer and then use that data
		nm.msg = msg
		s.output <- nm
	})
	if err != nil {
		s.err = err
		return s.output
	}
	if err := s.conn.Flush(); err != nil {
		s.err = err
		return s.output
	}
	if err := s.conn.LastError(); err != nil {
		s.err = err
		return s.output
	}
	return s.output
}

func (s *natsSubscriber) Err() error {
	return s.err
}

func (s *natsSubscriber) Stop() error {
	s.conn.Close()
	close(s.output)
	return nil
}
