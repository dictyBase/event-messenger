package commands

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/dictyBase/event-messenger/internal/message"
	"google.golang.org/grpc"

	gclient "github.com/dictyBase/event-messenger/internal/message/grpc-client"
	"github.com/dictyBase/event-messenger/internal/message/nats"

	cli "gopkg.in/urfave/cli.v1"
)

// CreateIssue creates a Github issue when a new stock order comes through
func CreateIssue(c *cli.Context) error {
	s, err := nats.NewSubscriber(c.String("nats-host"), c.String("nats-port"))
	if err != nil {
		log.Fatalf("cannot connect to nats for subscription %s\n", err)
	}
	// still need to add other grpc dialers (stock, annotation)
	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%s", c.String("order-grpc-host"), c.String("order-grpc-port")),
		grpc.WithInsecure(),
	)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("cannot connect to grpc server for order microservice %s", err),
			2,
		)
	}
	err = s.Start(
		"OrderService.*",
		gclient.NewOrderClient(conn),
	)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("cannot start the subscriber server %s", err),
			2,
		)
	}
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
