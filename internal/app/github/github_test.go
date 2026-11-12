package github

import (
	"errors"
	"io"
	"testing"

	F "github.com/IBM/fp-go/v2/function"
	"github.com/dictyBase/event-messenger/internal/datasource"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestToCLIExitPreservesMessageAndExitCode(t *testing.T) {
	t.Parallel()

	err := toCLIExit(errors.New("nats connection refused"))

	var exitErr cli.ExitCoder
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, exitCode, exitErr.ExitCode())
	require.Equal(t, "nats connection refused", exitErr.Error())
}

func TestCreateIssueTrackerBuildsTrackerFromSetupState(t *testing.T) {
	t.Parallel()

	lg := logrus.New()
	lg.SetOutput(io.Discard)

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "token", Value: "gh-test-token"},
			&cli.StringFlag{Name: "owner", Value: "dictyBase"},
			&cli.StringFlag{Name: "repository", Value: "Stock-Center-Orders"},
			&cli.StringFlag{Name: "openrouter-token", Value: ""},
			&cli.StringFlag{Name: "jev-model", Value: ""},
			&cli.IntFlag{Name: "strain-price", Value: 30},
			&cli.IntFlag{Name: "plasmid-price", Value: 15},
		},
	}

	state := F.Pipe3(
		IssueSetupState{},
		setupCommandLens.Set(cmd),
		setupLoggerLens.Set(logrus.NewEntry(lg)),
		setupSourcesLens.Set(&datasource.Sources{}),
	)

	tracker := createIssueTracker(state)
	require.NotNil(t, tracker)
}
