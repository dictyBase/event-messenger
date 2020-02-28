package gmail

import (
	"os"
	"os/signal"
	"syscall"

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
	shutdown(s, l)
	return nil
}

func setupGmail(c *cli.Context, logger *logrus.Entry) (message.GmailSubscriber, error) {
	s, err := nats.NewGmailSubscriber(c.String("nats-host"), c.String("nats-port"), logger)
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

func shutdown(r message.GmailSubscriber, logger *logrus.Entry) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	<-ch
	logger.Info("received kill signal")
	if err := r.Stop(); err != nil {
		logger.Fatalf("unable to close the subscription %s\n", err)
	}
	logger.Info("closed the connections gracefully")
}
