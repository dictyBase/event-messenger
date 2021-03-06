package github

import (
	"github.com/dictyBase/event-messenger/internal/datasource"
	gh "github.com/dictyBase/event-messenger/internal/issue-tracker/github"
	"github.com/dictyBase/event-messenger/internal/logger"
	"github.com/dictyBase/event-messenger/internal/message"
	"github.com/dictyBase/event-messenger/internal/message/nats"
	"github.com/dictyBase/event-messenger/internal/service"
	"github.com/urfave/cli"
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
	mc, err := service.ClientConn(c, []string{"stock", "user", "annotation"})
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
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
		return cli.NewExitError(err.Error(), 2)
	}
	l.Info("starting the Github issue creation subscriber messaging backend")
	message.Shutdown(s, l)
	return nil
}
