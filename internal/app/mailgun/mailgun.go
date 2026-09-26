package mailgun

import (
	"context"
	"fmt"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	L "github.com/IBM/fp-go/v2/optics/lens"
	"github.com/dictyBase/event-messenger/internal/datasource"
	"github.com/dictyBase/event-messenger/internal/logger"
	"github.com/dictyBase/event-messenger/internal/message"
	"github.com/dictyBase/event-messenger/internal/message/nats"
	email "github.com/dictyBase/event-messenger/internal/send-email"
	emailmg "github.com/dictyBase/event-messenger/internal/send-email/mailgun"
	"github.com/dictyBase/event-messenger/internal/service"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
)

const exitCode = 2

// EmailActionInput is the boundary input of the send-email action.
type EmailActionInput struct {
	Context context.Context
	Command *cli.Command
}

// EmailSetupState accumulates the subscriber, the data sources and the
// mailer as the setup pipeline runs.
type EmailSetupState struct {
	Context     context.Context
	Command     *cli.Command
	Logger      *logrus.Entry
	Subscriber  *nats.EmailSubscriber
	Sources     *datasource.Sources
	Publication *datasource.Publication
	Mailer      email.Handler
}

var (
	// setupLoggerLens focuses the assembled logger.
	setupLoggerLens = L.MakeLens(
		func(s EmailSetupState) *logrus.Entry { return s.Logger },
		func(s EmailSetupState, v *logrus.Entry) EmailSetupState {
			s.Logger = v
			return s
		},
	)

	// setupSubLens focuses the connected nats subscriber.
	setupSubLens = L.MakeLens(
		func(s EmailSetupState) *nats.EmailSubscriber { return s.Subscriber },
		func(s EmailSetupState, v *nats.EmailSubscriber) EmailSetupState {
			s.Subscriber = v
			return s
		},
	)

	// setupSourcesLens focuses the grpc data sources.
	setupSourcesLens = L.MakeLens(
		func(s EmailSetupState) *datasource.Sources { return s.Sources },
		func(s EmailSetupState, v *datasource.Sources) EmailSetupState {
			s.Sources = v
			return s
		},
	)

	// setupPubLens focuses the publication client.
	setupPubLens = L.MakeLens(
		func(s EmailSetupState) *datasource.Publication { return s.Publication },
		func(s EmailSetupState, v *datasource.Publication) EmailSetupState {
			s.Publication = v
			return s
		},
	)

	// setupMailerLens focuses the wired email handler.
	setupMailerLens = L.MakeLens(
		func(s EmailSetupState) email.Handler { return s.Mailer },
		func(s EmailSetupState, v email.Handler) EmailSetupState {
			s.Mailer = v
			return s
		},
	)
)

// toCLIExit maps a setup failure onto the action's exit code.
func toCLIExit(err error) error {
	msg := err.Error()

	return cli.Exit(msg, exitCode)
}

// RunSendEmail connects to nats and sends an email based on received
// stock order data.
func RunSendEmail(ctx context.Context, c *cli.Command) error {
	return F.Pipe12(
		EmailActionInput{Context: ctx, Command: c},
		seedSetupState,
		IOE.Bind(setupLoggerLens.Set, createLogger),
		IOE.Bind(setupSubLens.Set, createSubscriber),
		IOE.Bind(setupSourcesLens.Set, createSources),
		IOE.Bind(setupPubLens.Set, createPublicationSource),
		IOE.Let[error](setupMailerLens.Set, createMailer),
		IOE.Chain(startSubscription),
		IOE.ChainFirstIOK[error](logStartup),
		IOE.ChainFirstIOK[error](awaitShutdown),
		ioeutils.ToEither[error, EmailSetupState],
		E.MapLeft[EmailSetupState](toCLIExit),
		E.ToError[EmailSetupState],
	)
}

// seedSetupState lifts the boundary input into the setup state.
func seedSetupState(in EmailActionInput) IOE.IOEither[error, EmailSetupState] {
	return IOE.Of[error](EmailSetupState{
		Context: in.Context,
		Command: in.Command,
	})
}

// createLogger builds the action logger from the command flags.
func createLogger(s EmailSetupState) IOE.IOEither[error, *logrus.Entry] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*logrus.Entry, error) {
			return logger.NewLogger(s.Command)
		}),
		IOE.MapLeft[*logrus.Entry](func(err error) error {
			return fmt.Errorf("error creating logger: %w", err)
		}),
	)
}

// createSubscriber connects the nats subscriber.
func createSubscriber(
	s EmailSetupState,
) IOE.IOEither[error, *nats.EmailSubscriber] {
	host := s.Command.String("nats-host")
	port := s.Command.String("nats-port")

	return F.Pipe1(
		IOE.TryCatchError(func() (*nats.EmailSubscriber, error) {
			return nats.NewEmailSubscriber(host, port, s.Logger)
		}),
		IOE.MapLeft[*nats.EmailSubscriber](func(err error) error {
			return fmt.Errorf("error creating nats subscriber: %w", err)
		}),
	)
}

// createSources dials the grpc services and wraps them as data sources.
func createSources(s EmailSetupState) IOE.IOEither[error, *datasource.Sources] {
	services := []string{"stock", "annotation", "user"}

	return F.Pipe2(
		IOE.TryCatchError(func() (map[string]*grpc.ClientConn, error) {
			return service.ClientConn(s.Command, services)
		}),
		IOE.MapLeft[map[string]*grpc.ClientConn](func(err error) error {
			return fmt.Errorf("error connecting to grpc services: %w", err)
		}),
		IOE.Map[error](datasource.GrpcSources),
	)
}

// createPublicationSource builds the NCBI eUtils publication client.
func createPublicationSource(
	_ EmailSetupState,
) IOE.IOEither[error, *datasource.Publication] {
	pub := datasource.NewPublication()

	return F.Pipe1(
		pub,
		IOE.MapLeft[*datasource.Publication](func(err error) error {
			return fmt.Errorf("error setting up publication client: %w", err)
		}),
	)
}

// createMailer wires the mailgun email handler from the command flags.
func createMailer(s EmailSetupState) email.Handler {
	return emailmg.NewMailgunEmailer(&emailmg.EmailerParams{
		Sender:       s.Command.String("sender"),
		SenderName:   s.Command.String("name"),
		Domain:       s.Command.String("domain"),
		APIKey:       s.Command.String("apiKey"),
		EmailCC:      s.Command.String("cc"),
		StrainPrice:  s.Command.Int("strain-price"),
		PlasmidPrice: s.Command.Int("plasmid-price"),
		Logger:       s.Logger,
		Sources:      s.Sources,
		PubSource:    s.Publication,
	})
}

// rawStartSubscription subscribes the mailer to the subject.
func rawStartSubscription(
	s *nats.EmailSubscriber,
	subject string,
	mailer email.Handler,
) (IO.Void, error) {
	err := s.Start(subject, mailer)

	return struct{}{}, err
}

// startSubscription starts the subscription and keeps the setup state.
func startSubscription(
	s EmailSetupState,
) IOE.IOEither[error, EmailSetupState] {
	sub := s.Command.String("subject")

	return F.Pipe2(
		IOE.TryCatchError(func() (IO.Void, error) {
			return rawStartSubscription(s.Subscriber, sub, s.Mailer)
		}),
		IOE.MapLeft[IO.Void](func(err error) error {
			return fmt.Errorf("error starting email subscriber: %w", err)
		}),
		IOE.MapTo[error, IO.Void](s),
	)
}

// logStartup records that the subscriber is running. Named IO tap: the
// logger is a *logrus.Entry on the state and IO.Logf would bypass the
// configured logrus output.
func logStartup(s EmailSetupState) IO.IO[IO.Void] {
	return IO.FromImpure(func() {
		s.Logger.Info("starting the email sending subscriber backend")
	})
}

// awaitShutdown blocks until a kill signal arrives and closes the
// subscriber. This is a lifecycle effect, not logging.
func awaitShutdown(s EmailSetupState) IO.IO[IO.Void] {
	return IO.FromImpure(func() {
		message.Shutdown(s.Subscriber, s.Logger)
	})
}
