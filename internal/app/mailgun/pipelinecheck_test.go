package mailgun

import (
	"strings"
	"testing"

	pipelinecheck "github.com/dictyBase/fp-go-loom/pipelinecheck"
)

func TestPipelineContinuity(t *testing.T) {
	t.Parallel()

	cfg := pipelinecheck.Config{
		Roots: []string{"."},
		IsEntrypoint: func(name string) bool {
			return name == "RunSendEmail" || strings.HasPrefix(name, "run")
		},
		RequirePointFreeSeed:      true,
		RequirePointFreeBranching: true,
	}
	pipelinecheck.Require(t, cfg)

	rawCfg := pipelinecheck.Config{
		Roots: []string{
			".",
			"../../datasource",
			"../../send-email/mailgun",
		},
	}
	pipelinecheck.RequireNoNonRawTryCatchCallback(t, rawCfg)
	pipelinecheck.RequireNoIfErrInTryCatch(t, rawCfg)
	pipelinecheck.RequireBareIf(t, rawCfg)
}
