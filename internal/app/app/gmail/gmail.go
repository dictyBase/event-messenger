package gmail

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/dictyBase/event-messenger/internal/message"
	"github.com/dictyBase/event-messenger/internal/message/nats"
	gm "github.com/dictyBase/event-messenger/internal/send-email/gmail"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func getLogger(c *cli.Context) *logrus.Entry {
	log := logrus.New()
	log.Out = os.Stderr
	switch c.GlobalString("log-format") {
	case "text":
		log.Formatter = &logrus.TextFormatter{
			TimestampFormat: "02/Jan/2006:15:04:05",
		}
	case "json":
		log.Formatter = &logrus.JSONFormatter{
			TimestampFormat: "02/Jan/2006:15:04:05",
		}
	}
	l := c.GlobalString("log-level")
	switch l {
	case "debug":
		log.Level = logrus.DebugLevel
	case "warn":
		log.Level = logrus.WarnLevel
	case "error":
		log.Level = logrus.ErrorLevel
	case "fatal":
		log.Level = logrus.FatalLevel
	case "panic":
		log.Level = logrus.PanicLevel
	}
	return logrus.NewEntry(log)
}

// RunSendEmail connects to nats and sends an email based on received stock order data.
func RunSendEmail(c *cli.Context) error {
	l := getLogger(c)
	s, err := nats.NewGmailSubscriber(c.String("nats-host"), c.String("nats-port"), l)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	cl, err := gm.GetGmailClient(c)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	g := gm.NewEmailSender(c.String("secret"), c.String("reply-to"), c.String("send-to"), cl, l)
	err = s.Start(c.String("subject"), g)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	l.Info("starting the email sending subscriber backend")
	shutdown(s, l)
	return nil
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
