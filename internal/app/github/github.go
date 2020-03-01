package github

import (
	"fmt"

	gh "github.com/dictyBase/event-messenger/internal/issue-tracker/github"
	"github.com/dictyBase/event-messenger/internal/logger"
	"github.com/dictyBase/event-messenger/internal/message"
	"github.com/dictyBase/event-messenger/internal/message/nats"
	"github.com/dictyBase/go-genproto/dictybaseapis/annotation"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

// RunCreateIssue connects to nats and creates a GitHub issue based on received order data.
func RunCreateIssue(c *cli.Context) error {
	l, err := logger.NewLogger(c)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	s, err := nats.NewGithubSubscriber(
		c.String("nats-host"),
		c.String("nats-port"),
		l,
	)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	sconn, aconn, err := serviceClients(c)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	g := gh.NewIssueCreator(&gh.IssueParams{
		Token:       c.String("token"),
		Owner:       c.String("owner"),
		Repository:  c.String("repository"),
		AnnoClient:  annotation.NewTaggedAnnotationServiceClient(aconn),
		StockClient: stock.NewStockServiceClient(sconn),
		Logger:      l,
	})
	err = s.Start(c.String("subject"), g)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	l.Info("starting the Github issue creation subscriber messaging backend")
	message.Shutdown(s, l)
	return nil
}

func serviceClients(c *cli.Context) (*grpc.ClientConn, *grpc.ClientConn, error) {
	sconn, err := grpc.Dial(
		fmt.Sprintf("%s:%s", c.String("stock-grpc-host"), c.String("stock-grpc-port")),
		grpc.WithInsecure(),
	)
	if err != nil {
		return sconn, sconn, fmt.Errorf("error in connecting to stock service %s", err)
	}
	aconn, err := grpc.Dial(
		fmt.Sprintf("%s:%s", c.String("annotation-grpc-host"), c.String("annotation-grpc-port")),
		grpc.WithInsecure(),
	)
	if err != nil {
		return aconn, aconn, fmt.Errorf("error in connecting to annotation service %s", err)
	}
	return sconn, aconn, nil
}
