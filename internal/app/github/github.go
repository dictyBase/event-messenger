package github

import (
	"github.com/dictyBase/event-messenger/internal/datasource"
	gh "github.com/dictyBase/event-messenger/internal/issue-tracker/github"
	"github.com/dictyBase/event-messenger/internal/logger"
	"github.com/dictyBase/event-messenger/internal/message"
	"github.com/dictyBase/event-messenger/internal/message/nats"
	"github.com/dictyBase/event-messenger/internal/service"
	"github.com/dictyBase/go-genproto/dictybaseapis/annotation"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/dictyBase/go-genproto/dictybaseapis/user"
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
	mc, err := service.ServiceClients(c, []string{"stock", "user", "annotation"})
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	g := gh.NewIssueCreator(&gh.IssueParams{
		Logger:     l,
		Token:      c.String("token"),
		Owner:      c.String("owner"),
		Repository: c.String("repository"),
		AnnoSource: &datasource.Annotation{
			Client: annotation.NewTaggedAnnotationServiceClient(mc["annotation"]),
		},
		StockSource: &datasource.Stock{
			Client: stock.NewStockServiceClient(mc["stock"]),
		},
		UserSource: &datasource.User{
			Client: user.NewUserServiceClient(mc["user"]),
		},
	})
	err = s.Start(c.String("subject"), g)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	l.Info("starting the Github issue creation subscriber messaging backend")
	message.Shutdown(s, l)
	return nil
}
