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

// IssueActionInput is the boundary input of the gh-issue action.
type IssueActionInput struct {
	Context context.Context
	Command *cli.Command
}

// IssueSetupState accumulates the logger, subscriber, data sources, and
// GitHub issue tracker as the setup pipeline runs.
type IssueSetupState struct {
	Context    context.Context
	Command    *cli.Command
	Logger     *logrus.Entry
	Subscriber *nats.GithubSubscriber
	Sources    *datasource.Sources
	Tracker    issue.Tracker
}

var (
	actionContextLens = L.MakeLens(
		func(in IssueActionInput) context.Context { return in.Context },
		func(in IssueActionInput, v context.Context) IssueActionInput {
			in.Context = v
			return in
		},
	)

	actionCommandLens = L.MakeLens(
		func(in IssueActionInput) *cli.Command { return in.Command },
		func(in IssueActionInput, v *cli.Command) IssueActionInput {
			in.Command = v
			return in
		},
	)

	setupContextLens = L.MakeLens(
		func(s IssueSetupState) context.Context { return s.Context },
		func(s IssueSetupState, v context.Context) IssueSetupState {
			s.Context = v
			return s
		},
	)

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

	paramsLoggerLens = L.MakeLens(
		func(p gh.IssueParams) *logrus.Entry { return p.Logger },
		func(p gh.IssueParams, v *logrus.Entry) gh.IssueParams {
			p.Logger = v
			return p
		},
	)

	paramsTokenLens = L.MakeLens(
		func(p gh.IssueParams) string { return p.Token },
		func(p gh.IssueParams, v string) gh.IssueParams {
			p.Token = v
			return p
		},
	)

	paramsOwnerLens = L.MakeLens(
		func(p gh.IssueParams) string { return p.Owner },
		func(p gh.IssueParams, v string) gh.IssueParams {
			p.Owner = v
			return p
		},
	)

	paramsRepoLens = L.MakeLens(
		func(p gh.IssueParams) string { return p.Repository },
		func(p gh.IssueParams, v string) gh.IssueParams {
			p.Repository = v
			return p
		},
	)

	paramsOpenRouterLens = L.MakeLens(
		func(p gh.IssueParams) string { return p.OpenRouterToken },
		func(p gh.IssueParams, v string) gh.IssueParams {
			p.OpenRouterToken = v
			return p
		},
	)

	paramsJevModelLens = L.MakeLens(
		func(p gh.IssueParams) string { return p.JevModel },
		func(p gh.IssueParams, v string) gh.IssueParams {
			p.JevModel = v
			return p
		},
	)

	paramsSourcesLens = L.MakeLens(
		func(p gh.IssueParams) *datasource.Sources { return p.Sources },
		func(p gh.IssueParams, v *datasource.Sources) gh.IssueParams {
			p.Sources = v
			return p
		},
	)

	paramsStrainPriceLens = L.MakeLens(
		func(p gh.IssueParams) int { return p.StrainPrice },
		func(p gh.IssueParams, v int) gh.IssueParams {
			p.StrainPrice = v
			return p
		},
	)

	paramsPlasmidPriceLens = L.MakeLens(
		func(p gh.IssueParams) int { return p.PlasmidPrice },
		func(p gh.IssueParams, v int) gh.IssueParams {
			p.PlasmidPrice = v
			return p
		},
	)
)

func toCLIExit(err error) error {
	msg := err.Error()

	return cli.Exit(msg, exitCode)
}

// RunCreateIssue connects to nats and creates a GitHub issue based on
// received order data.
func RunCreateIssue(ctx context.Context, c *cli.Command) error {
	return F.Pipe11(
		IssueActionInput{Context: ctx, Command: c},
		seedIssueSetupState,
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

func seedIssueSetupState(
	in IssueActionInput,
) IOE.IOEither[error, IssueSetupState] {
	return F.Pipe3(
		IssueSetupState{},
		setupContextLens.Set(actionContextLens.Get(in)),
		setupCommandLens.Set(actionCommandLens.Get(in)),
		IOE.Of[error],
	)
}

func createIssueLogger(s IssueSetupState) IOE.IOEither[error, *logrus.Entry] {
	cmd := setupCommandLens.Get(s)

	return F.Pipe1(
		IOE.TryCatchError(func() (*logrus.Entry, error) {
			return logger.NewLogger(cmd)
		}),
		IOE.MapLeft[*logrus.Entry](func(err error) error {
			return fmt.Errorf("error creating logger: %w", err)
		}),
	)
}

func createIssueSubscriber(
	s IssueSetupState,
) IOE.IOEither[error, *nats.GithubSubscriber] {
	cmd := setupCommandLens.Get(s)
	logEntry := setupLoggerLens.Get(s)
	host := cmd.String("nats-host")
	port := cmd.String("nats-port")

	return F.Pipe1(
		IOE.TryCatchError(func() (*nats.GithubSubscriber, error) {
			return nats.NewGithubSubscriber(host, port, logEntry)
		}),
		IOE.MapLeft[*nats.GithubSubscriber](func(err error) error {
			return fmt.Errorf("error creating github nats subscriber: %w", err)
		}),
	)
}

func createIssueSources(
	s IssueSetupState,
) IOE.IOEither[error, *datasource.Sources] {
	cmd := setupCommandLens.Get(s)
	services := []string{"stock", "annotation"}

	return F.Pipe2(
		IOE.TryCatchError(func() (map[string]*grpc.ClientConn, error) {
			return service.ClientConn(cmd, services)
		}),
		IOE.MapLeft[map[string]*grpc.ClientConn](func(err error) error {
			return fmt.Errorf("error connecting to grpc services: %w", err)
		}),
		IOE.Map[error](datasource.GrpcSources),
	)
}

func createIssueTracker(s IssueSetupState) issue.Tracker {
	cmd := setupCommandLens.Get(s)

	return F.Pipe10(
		gh.IssueParams{},
		paramsLoggerLens.Set(setupLoggerLens.Get(s)),
		paramsTokenLens.Set(cmd.String("token")),
		paramsOwnerLens.Set(cmd.String("owner")),
		paramsRepoLens.Set(cmd.String("repository")),
		paramsOpenRouterLens.Set(cmd.String("openrouter-token")),
		paramsJevModelLens.Set(cmd.String("jev-model")),
		paramsSourcesLens.Set(setupSourcesLens.Get(s)),
		paramsStrainPriceLens.Set(cmd.Int("strain-price")),
		paramsPlasmidPriceLens.Set(cmd.Int("plasmid-price")),
		func(p gh.IssueParams) issue.Tracker {
			return gh.NewIssueCreator(new(p))
		},
	)
}

func rawStartIssueSubscription(
	sub *nats.GithubSubscriber,
	subject string,
	tracker issue.Tracker,
) (IO.Void, error) {
	err := sub.Start(subject, tracker)

	return struct{}{}, err
}

func startIssueSubscription(
	s IssueSetupState,
) IOE.IOEither[error, IssueSetupState] {
	subject := setupCommandLens.Get(s).String("subject")
	sub := setupSubLens.Get(s)
	tracker := setupTrackerLens.Get(s)

	return F.Pipe2(
		IOE.TryCatchError(func() (IO.Void, error) {
			return rawStartIssueSubscription(sub, subject, tracker)
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
