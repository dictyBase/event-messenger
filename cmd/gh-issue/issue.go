package commands

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/dictyBase/event-messenger/internal/message"

	"github.com/dictyBase/event-messenger/internal/message/nats"

	cli "gopkg.in/urfave/cli.v1"
)

// CreateIssue creates a Github issue when a new stock order comes through
func CreateIssue(c *cli.Context) error {
	// connect subscriber to nats server
	s, err := nats.NewSubscriber(c.String("nats-host"), c.String("nats-port"))
	if err != nil {
		log.Fatalf("cannot connect to nats for subscription %s\n", err)
	}

	// need to add grpc dialers
	// need to call Start function

	logger := getLogger(c)
	logger.Info("starting the Github issue creation subscriber messaging backend")
	shutdown(s, logger)
	return nil
}

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

func shutdown(r message.NatsSubscriber, logger *logrus.Entry) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	<-ch
	logger.Info("received kill signal")
	if err := r.Stop(); err != nil {
		logger.Fatalf("unable to close the subscription %s\n", err)
	}
	logger.Info("closed the connections gracefully")
}
