package github

import (
	"strings"
	"testing"

	pipelinecheck "github.com/dictyBase/fp-go-loom/pipelinecheck"
)

func TestPipelineContinuity(t *testing.T) {
	t.Parallel()

	cfg := pipelinecheck.Config{
		Roots: []string{
			".",
			"../../issue-tracker/github",
		},
		IsEntrypoint: func(name string) bool {
			return name == "RunCreateIssue" || strings.HasPrefix(name, "run")
		},
		TerminalFunctions: []string{
			"postIssue",
			"executePostIssue",
			"rawStartIssueSubscription",
			"startIssueSubscription",
		},
		RequirePointFreeSeed:      true,
		RequirePointFreeBranching: true,
		RequireTrivialDelegate:    true,
	}
	pipelinecheck.Require(t, cfg)
	pipelinecheck.RequireBareIf(t, cfg)
	pipelinecheck.RequireNoIfErrInTryCatch(t, cfg)
	pipelinecheck.RequireNoNonRawTryCatchCallback(t, cfg)
	pipelinecheck.RequireSafePrintf(t, cfg)
	pipelinecheck.RequireNoRedundantFold(t, cfg)
	pipelinecheck.RequireNoHandRolledFileJoinInArrayMap(t, cfg)
	pipelinecheck.RequireTerminalEffect(t, cfg)
	pipelinecheck.RequireTrivialDelegate(t, cfg)
	pipelinecheck.RequireNoImperativeLoop(t, cfg)
	pipelinecheck.RequireNoImperativeBranch(t, cfg)
	pipelinecheck.RequireNoDirectApply(t, cfg)
}
