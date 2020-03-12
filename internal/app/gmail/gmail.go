package gmail

import (
	"github.com/dictyBase/event-messenger/internal/logger"
	"github.com/dictyBase/event-messenger/internal/message"
	"github.com/dictyBase/event-messenger/internal/message/nats"
	gm "github.com/dictyBase/event-messenger/internal/send-email/gmail"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// RunSendEmail connects to nats and sends an email based on received stock order data.
func RunSendEmail(c *cli.Context) error {
	l, err := logger.NewLogger(c)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	s, err := setupGmail(c, l)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	l.Info("starting the email sending subscriber backend")
	message.Shutdown(s, l)
	return nil
}

func setupGmail(c *cli.Context, logger *logrus.Entry) (*nats.NatsEmailSubscriber, error) {
	s, err := nats.NewEmailSubscriber(c.String("nats-host"), c.String("nats-port"), logger)
	if err != nil {
		return s, err
	}
	cl, err := gm.GetGmailClient(c)
	if err != nil {
		return s, err
	}
	g := gm.NewEmailSender(c.String("secret"), c.String("reply-to"), c.String("send-to"), cl, logger)
	return s, s.Start(c.String("subject"), g)
}
