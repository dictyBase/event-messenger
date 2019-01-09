package commands

import (
	"fmt"
	"log"

	"github.com/dictyBase/event-messenger/message/nats"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"

	gnats "github.com/nats-io/go-nats"
	cli "gopkg.in/codegangsta/cli.v1"
)

// ValidateServerArgs validates that the necessary flags are not missing
func ValidateServerArgs(c *cli.Context) error {
	for _, p := range []string{"repository", "owner", "gh-token", "nats-host",
		"nats-port"} {
		if len(c.String(p)) == 0 {
			return cli.NewExitError(
				fmt.Sprintf("argument %s is missing", p),
				2,
			)
		}
	}
	return nil
}

// CreateIssue creates a Github issue when a new order comes through
func CreateIssue(c *cli.Context) error {
	if err := ValidateServerArgs(c); err != nil {
		log.Fatal(err)
	}

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
