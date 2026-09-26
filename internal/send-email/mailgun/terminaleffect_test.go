package mailgun

import (
	"testing"

	pipelinecheck "github.com/dictyBase/fp-go-loom/pipelinecheck"
)

func TestTerminalEffect(t *testing.T) {
	t.Parallel()

	pipelinecheck.RequireTerminalEffect(t, pipelinecheck.Config{
		Roots:             []string{"."},
		TerminalFunctions: []string{"Send"},
	})
}
