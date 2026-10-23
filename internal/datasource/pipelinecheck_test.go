package datasource

import (
	"testing"

	pipelinecheck "github.com/dictyBase/fp-go-loom/pipelinecheck"
)

func TestPipelineContinuity(t *testing.T) {
	t.Parallel()

	cfg := pipelinecheck.Config{
		Roots:                     []string{"."},
		RequirePointFreeBranching: true,
		RequireTrivialDelegate:    true,
	}
	pipelinecheck.Require(t, cfg)

	pipelinecheck.RequireNoNonRawTryCatchCallback(t, cfg)
	pipelinecheck.RequireNoIfErrInTryCatch(t, cfg)
	pipelinecheck.RequireBareIf(t, cfg)
}
