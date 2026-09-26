package mailgun

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestToCLIExit(t *testing.T) {
	t.Parallel()

	err := toCLIExit(errors.New("boom"))

	var exitErr cli.ExitCoder
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, exitCode, exitErr.ExitCode())
	require.Equal(t, "boom", exitErr.Error())
}
