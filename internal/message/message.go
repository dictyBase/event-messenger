package message

import (
	"os"
	"os/signal"
	"syscall"

	gnats "github.com/nats-io/go-nats"
	"github.com/sirupsen/logrus"
)

// Subscriber is a generic interface for any kind of subscriber
type Subscriber interface {
	// Stop will initiate a graceful shutdown of the subscriber connection.
	Stop() error
}

func Shutdown(r Subscriber, logger *logrus.Entry) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	<-ch
	logger.Info("received kill signal")
	if err := r.Stop(); err != nil {
		logger.Fatalf("unable to close the subscription %s\n", err)
	}
	logger.Info("closed the connections gracefully")
}

func HandleConnection(econn *gnats.EncodedConn, err error) error {
	if err != nil {
		return err
	}
	if err := econn.Flush(); err != nil {
		return err
	}
	return econn.LastError()
}
