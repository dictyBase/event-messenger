package mailgun

import (
	"github.com/dictyBase/event-messenger/internal/datasource"
	"github.com/dictyBase/event-messenger/internal/logger"
	"github.com/dictyBase/event-messenger/internal/message"
	"github.com/dictyBase/event-messenger/internal/message/nats"
	mg "github.com/dictyBase/event-messenger/internal/send-email/mailgun"
	"github.com/dictyBase/event-messenger/internal/service"
	"github.com/dictyBase/go-genproto/dictybaseapis/annotation"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/dictyBase/go-genproto/dictybaseapis/user"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// RunSendEmail connects to nats and sends an email based on received stock order data.
func RunSendEmail(c *cli.Context) error {
	l, err := logger.NewLogger(c)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	s, err := setupEmail(c, l)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	l.Info("starting the email sending subscriber backend")
	message.Shutdown(s, l)
	return nil
}

func setupEmail(c *cli.Context, logger *logrus.Entry) (*nats.NatsEmailSubscriber, error) {
	s, err := nats.NewEmailSubscriber(c.String("nats-host"), c.String("nats-port"), logger)
	if err != nil {
		return s, err
	}
	mc, err := service.ServiceClients(c, []string{"stock", "annotation", "user"})
	if err != nil {
		return s, err
	}
	mailer := mg.NewMailgunEmailer(&mg.EmailerParams{
		Sender:       c.String("sender"),
		SenderName:   c.String("name"),
		Domain:       c.String("domain"),
		ApiKey:       c.String("apiKey"),
		StrainPrice:  c.Int("strain-price"),
		PlasmidPrice: c.Int("plasmid-price"),
		Logger:       logger,
		AnnoSource: &datasource.Annotation{
			Client: annotation.NewTaggedAnnotationServiceClient(mc["annotation"]),
		},
		StockSource: &datasource.Stock{
			Client: stock.NewStockServiceClient(mc["stock"]),
		},
		UserSource: &datasource.User{
			Client: user.NewUserServiceClient(mc["user"]),
		},
		PubSource: datasource.NewPublication(c.String("publication-api")),
	})
	return s, s.Start(c.String("subject"), mailer)
}
