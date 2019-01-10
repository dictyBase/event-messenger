package commands

import (
	"log"

	"github.com/dictyBase/event-messenger/internal/message/nats"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"

	gnats "github.com/nats-io/go-nats"
	cli "gopkg.in/urfave/cli.v1"
)

// CreateIssue creates a Github issue when a new order comes through
func CreateIssue(c *cli.Context) error {

	return nil
}

// natsTest is just an example function for connecting a subscriber
// to nats server and then opening a channel
func natsTest(subj string, ord *order.Order) string {
	// connect subscriber to nats server
	s, err := nats.NewSubscriber(gnats.DefaultURL) // replace with CLI string
	if err != nil {
		log.Fatalf("cannot connect to nats for subscription %s\n", err)
	}
	// start server to communicate using a channel
	sc := s.Start(subj, ord)
	if err := s.Err(); err != nil {
		log.Fatalf("could not start subscription %s\n", err)
	}
	// async, Start function is triggered
	msg := <-sc
	if err := s.Stop(); err != nil {
		log.Fatalf("could not stop the subscription %s\n", err)
	}
	return string(msg.Message())
}
