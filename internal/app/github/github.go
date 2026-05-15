package github

import (
	"context"

	"github.com/dictyBase/event-messenger/internal/datasource"
	gh "github.com/dictyBase/event-messenger/internal/issue-tracker/github"
	"github.com/dictyBase/event-messenger/internal/logger"
	"github.com/dictyBase/event-messenger/internal/message"
	"github.com/dictyBase/event-messenger/internal/message/nats"
	"github.com/dictyBase/event-messenger/internal/service"
	"github.com/urfave/cli/v3"
)

const exitCode = 2

// RunCreateIssue connects to nats and creates a GitHub issue based on received order data.
func RunCreateIssue(_ context.Context, c *cli.Command) error {
	l, err := logger.NewLogger(c)
	if err != nil {
		return cli.Exit(err.Error(), exitCode)
	}

	s, err := nats.NewGithubSubscriber(
		c.String("nats-host"),
		c.String("nats-port"),
		l,
	)
	if err != nil {
		return cli.Exit(err.Error(), exitCode)
	}

	mc, err := service.ClientConn(c, []string{"stock", "user", "annotation"})
	if err != nil {
		return cli.Exit(err.Error(), exitCode)
	}

	g := gh.NewIssueCreator(&gh.IssueParams{
		Logger:       l,
		Token:        c.String("token"),
		Owner:        c.String("owner"),
		Repository:   c.String("repository"),
		Sources:      datasource.GrpcSources(mc),
		StrainPrice:  c.Int("strain-price"),
		PlasmidPrice: c.Int("plasmid-price"),
	})

	err = s.Start(c.String("subject"), g)
	if err != nil {
		return cli.Exit(err.Error(), exitCode)
	}

	l.Info("starting the Github issue creation subscriber messaging backend")
	message.Shutdown(s, l)

	return nil
}
