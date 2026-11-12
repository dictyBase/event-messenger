package github

import (
	"context"
	"fmt"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	L "github.com/IBM/fp-go/v2/optics/lens"
	"github.com/dictyBase/event-messenger/internal/datasource"
	issue "github.com/dictyBase/event-messenger/internal/issue-tracker"
	gh "github.com/dictyBase/event-messenger/internal/issue-tracker/github"
	"github.com/dictyBase/event-messenger/internal/logger"
	"github.com/dictyBase/event-messenger/internal/message"
	"github.com/dictyBase/event-messenger/internal/message/nats"
	"github.com/dictyBase/event-messenger/internal/service"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
)

const exitCode = 2

// IssueSetupState accumulates the logger, subscriber, data sources, and
// GitHub issue tracker as the setup pipeline runs.
type IssueSetupState struct {
	Command    *cli.Command
	Logger     *logrus.Entry
	Subscriber *nats.GithubSubscriber
	Sources    *datasource.Sources
	Tracker    issue.Tracker
}

var (
	issueRequiredServices = []string{"stock", "annotation"}

	setupCommandLens = L.MakeLens(
		func(s IssueSetupState) *cli.Command { return s.Command },
		func(s IssueSetupState, v *cli.Command) IssueSetupState {
			s.Command = v
			return s
		},
	)

	setupLoggerLens = L.MakeLens(
		func(s IssueSetupState) *logrus.Entry { return s.Logger },
		func(s IssueSetupState, v *logrus.Entry) IssueSetupState {
			s.Logger = v
			return s
		},
	)

	setupSubLens = L.MakeLens(
		func(s IssueSetupState) *nats.GithubSubscriber { return s.Subscriber },
		func(s IssueSetupState, v *nats.GithubSubscriber) IssueSetupState {
			s.Subscriber = v
			return s
		},
	)

	setupSourcesLens = L.MakeLens(
		func(s IssueSetupState) *datasource.Sources { return s.Sources },
		func(s IssueSetupState, v *datasource.Sources) IssueSetupState {
			s.Sources = v
			return s
		},
	)

	setupTrackerLens = L.MakeLens(
		func(s IssueSetupState) issue.Tracker { return s.Tracker },
		func(s IssueSetupState, v issue.Tracker) IssueSetupState {
			s.Tracker = v
			return s
		},
	)
)

func toCLIExit(err error) error {
	msg := err.Error()

	return cli.Exit(msg, exitCode)
}

// RunCreateIssue connects to nats and creates a GitHub issue based on
// received order data.
func RunCreateIssue(_ context.Context, c *cli.Command) error {
	return F.Pipe12(
		IssueSetupState{},
		setupCommandLens.Set(c),
		IOE.Of[error],
		IOE.Bind(setupLoggerLens.Set, createIssueLogger),
		IOE.Bind(setupSubLens.Set, createIssueSubscriber),
		IOE.Bind(setupSourcesLens.Set, createIssueSources),
		IOE.Let[error](setupTrackerLens.Set, createIssueTracker),
		IOE.Chain(startIssueSubscription),
		IOE.ChainFirstIOK[error](logIssueStartup),
		IOE.ChainFirstIOK[error](awaitIssueShutdown),
		ioeutils.ToEither[error, IssueSetupState],
		E.MapLeft[IssueSetupState](toCLIExit),
		E.ToError[IssueSetupState],
	)
}

func createIssueLogger(s IssueSetupState) IOE.IOEither[error, *logrus.Entry] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*logrus.Entry, error) {
			return logger.NewLogger(setupCommandLens.Get(s))
		}),
		IOE.MapLeft[*logrus.Entry](func(err error) error {
			return fmt.Errorf("error creating logger: %w", err)
		}),
	)
}

func createIssueSubscriber(
	s IssueSetupState,
) IOE.IOEither[error, *nats.GithubSubscriber] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*nats.GithubSubscriber, error) {
			return nats.NewGithubSubscriber(
				setupCommandLens.Get(s).String("nats-host"),
				setupCommandLens.Get(s).String("nats-port"),
				setupLoggerLens.Get(s),
			)
		}),
		IOE.MapLeft[*nats.GithubSubscriber](func(err error) error {
			return fmt.Errorf("error creating github nats subscriber: %w", err)
		}),
	)
}

func createIssueSources(
	s IssueSetupState,
) IOE.IOEither[error, *datasource.Sources] {
	return F.Pipe2(
		IOE.TryCatchError(func() (map[string]*grpc.ClientConn, error) {
			return service.ClientConn(
				setupCommandLens.Get(s),
				issueRequiredServices,
			)
		}),
		IOE.MapLeft[map[string]*grpc.ClientConn](func(err error) error {
			return fmt.Errorf("error connecting to grpc services: %w", err)
		}),
		IOE.Map[error](datasource.GrpcSources),
	)
}

func createIssueTracker(s IssueSetupState) issue.Tracker {
	return F.Pipe10(
		&gh.IssueParams{},
		F.Pipe2(s, setupLoggerLens.Get, gh.IssueParamsLoggerLens.Set),
		F.Pipe3(
			s,
			setupCommandLens.Get,
			F.Bind2nd((*cli.Command).String, "token"),
			gh.IssueParamsTokenLens.Set,
		),
		F.Pipe3(
			s,
			setupCommandLens.Get,
			F.Bind2nd((*cli.Command).String, "owner"),
			gh.IssueParamsOwnerLens.Set,
		),
		F.Pipe3(
			s,
			setupCommandLens.Get,
			F.Bind2nd((*cli.Command).String, "repository"),
			gh.IssueParamsRepoLens.Set,
		),
		F.Pipe3(
			s,
			setupCommandLens.Get,
			F.Bind2nd((*cli.Command).String, "openrouter-token"),
			gh.IssueParamsOpenRouterLens.Set,
		),
		F.Pipe3(
			s,
			setupCommandLens.Get,
			F.Bind2nd((*cli.Command).String, "jev-model"),
			gh.IssueParamsJevModelLens.Set,
		),
		F.Pipe2(s, setupSourcesLens.Get, gh.IssueParamsSourcesLens.Set),
		F.Pipe3(
			s,
			setupCommandLens.Get,
			F.Bind2nd((*cli.Command).Int, "strain-price"),
			gh.IssueParamsStrainPriceLens.Set,
		),
		F.Pipe3(
			s,
			setupCommandLens.Get,
			F.Bind2nd((*cli.Command).Int, "plasmid-price"),
			gh.IssueParamsPlasmidPriceLens.Set,
		),
		gh.NewIssueCreator,
	)
}

func rawStartIssueSubscription(s IssueSetupState) (IO.Void, error) {
	err := setupSubLens.Get(s).Start(
		setupCommandLens.Get(s).String("subject"),
		setupTrackerLens.Get(s),
	)

	return struct{}{}, err
}

func startIssueSubscription(
	s IssueSetupState,
) IOE.IOEither[error, IssueSetupState] {
	return F.Pipe2(
		IOE.TryCatchError(func() (IO.Void, error) {
			return rawStartIssueSubscription(s)
		}),
		IOE.MapLeft[IO.Void](func(err error) error {
			return fmt.Errorf("error starting github subscriber: %w", err)
		}),
		IOE.MapTo[error, IO.Void](s),
	)
}

func logIssueStartup(s IssueSetupState) IO.IO[IO.Void] {
	return IO.FromImpure(func() {
		setupLoggerLens.Get(s).Info(
			"starting the Github issue creation subscriber messaging backend",
		)
	})
}

func awaitIssueShutdown(s IssueSetupState) IO.IO[IO.Void] {
	return IO.FromImpure(func() {
		message.Shutdown(setupSubLens.Get(s), setupLoggerLens.Get(s))
	})
}
