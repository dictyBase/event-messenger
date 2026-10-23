# Pipelinecheck Mechanical Gate Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enforce the house fp-go pipeline shape mechanically in every
package of event-messenger with `pipelinecheck` from
`github.com/dictyBase/fp-go-loom/pipelinecheck`, so violations fail
`go test` instead of relying on review.

**Architecture:** Each package drops a thin `pipelinecheck_test.go`
that parses the package (and, where useful, sibling packages) and fails
the test on any violation. No production code changes. The existing
`internal/app/mailgun/pipelinecheck_test.go` is the in-repo template.

**Tech Stack:** Go, `github.com/dictyBase/fp-go-loom/pipelinecheck`,
testify not required. Lint gates: `golangci-lint fmt`,
`golangci-lint run ./...`, `scripts/check-gopls.sh` when present.

**Spec:** This plan implements layer 4 of the "make every project
follow the literature pattern" decision. Reference implementation:
`internal/datasource/publication.go` (pattern source) and
`internal/app/mailgun/pipelinecheck_test.go` (gate template).

## Global Constraints

- House stack only: `Option`, `Either`, `IO`, `IOEither`; alias `IOE` =
  `ioeither` forever.
- `fp-go-loom` must be upgraded to `v0.0.0-20261017160436-45e4f096bd46`
  or newer. The pinned version `20260907153825` predates rule 11
  (`RequireTrivialDelegate`) and the trivial-delegate directive.
- Opt-out directives need a non-empty reason:
  `// fp-go:allow-pipe1-handoff <reason>`,
  `// fp-go:allow-applied-seed <reason>`,
  `// fp-go:allow-bare-if <reason>`,
  `// fp-go:allow-trivial-delegate <reason>`.
- Every new test uses `t.Parallel()`.
- Do not touch production `.go` files except to fix violations that the
  gate exposes; fix, do not add opt-outs.
- Commits: small Conventional Commits, one per task where possible.
- Do not push. Leave branches for the user.

## Current State (verified 2026-10-22)

| Package | run-style entrypoints | Has gate |
|---|---|---|
| `internal/app/mailgun` | `RunSendEmail` | Yes (lacks `RequireTrivialDelegate`) |
| `internal/app/github` | `RunCreateIssue` | No |
| `internal/app/webhook` | `RunOntoServer` | No |
| `internal/datasource` | none | No (only reached by mailgun raw checks) |
| `internal/cmd` | flag/command builders, not pipelines | No (out of scope) |
| `cmd/event-messenger` | `main` only | No (out of scope) |

## Review Focus

1. A future `run*`/`Run*` handler that pre-applies its seed
   (`F.PipeN(seedX(ctx, c), ...)`) — the point-free seed rule must fail
   the test. Pinned by the violation-fixture step in Task 6.
2. A `F.Pipe1(seed, namedFn)` one-step handoff in an entrypoint — the
   handoff rule must fail the test. Pinned in Task 6.
3. A bare `if` statement inside an fp-go pipeline callback — the
   bare-if rule must fail the test. Pinned in Task 6.
4. A TryCatch callback with hand-rolled `if err != nil` instead of
   `MapLeft` — `RequireNoIfErrInTryCatch` must fail the test. Pinned in
   Task 6.
5. A single-statement pass-through wrapper referenced from a pipe —
   `RequireTrivialDelegate` must fail the test. Pinned in Task 6.

---

### Task 1: Upgrade fp-go-loom

**Files:**
- Modify: `go.mod`, `go.sum`

**Interfaces:**
- Produces: `pipelinecheck.RequireTrivialDelegate` and
  `pipelinecheck.Config.RequireTrivialDelegate` available for later
  tasks.

- [ ] **Step 1: Upgrade the dependency**

```sh
go get github.com/dictyBase/fp-go-loom@45e4f096bd46
go mod tidy
```

- [ ] **Step 2: Verify version and build**

Run: `go list -m github.com/dictyBase/fp-go-loom`
Expected: `github.com/dictyBase/fp-go-loom v0.0.0-20261017160436-45e4f096bd46`

Run: `go build ./...`
Expected: exit 0.

- [ ] **Step 3: Baseline tests still pass**

Run: `go test ./...`
Expected: PASS everywhere. If anything fails, stop and report the
pre-existing breakage before continuing.

- [ ] **Step 4: Commit**

```sh
git add go.mod go.sum
git commit -m "chore: bump fp-go-loom for pipelinecheck trivial-delegate rule"
```

### Task 2: Gate internal/datasource

**Files:**
- Create: `internal/datasource/pipelinecheck_test.go`

**Interfaces:**
- Consumes: `pipelinecheck.Require`, `pipelinecheck.RequireNoNonRawTryCatchCallback`,
  `pipelinecheck.RequireNoIfErrInTryCatch`, `pipelinecheck.RequireBareIf`,
  `pipelinecheck.Config{Roots, RequirePointFreeBranching, RequireTrivialDelegate}`.
- Produces: package-level regression gate. `datasource` has no `run*`
  entrypoints today, so the handoff/applied-seed rules are no-ops here;
  they still run as free insurance for future entrypoints.

- [ ] **Step 1: Write the test**

```go
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
```

- [ ] **Step 2: Run it**

Run: `go test ./internal/datasource -run TestPipelineContinuity -v`
Expected: PASS. The code already follows the style. If a violation
appears, fix the flagged code in `publication.go` per the violation
message, then rerun. Prefer fixing over adding an opt-out.

- [ ] **Step 3: Commit**

```sh
git add internal/datasource/pipelinecheck_test.go
git commit -m "test(datasource): add pipelinecheck mechanical gate"
```

### Task 3: Gate internal/app/github

**Files:**
- Create: `internal/app/github/pipelinecheck_test.go`

**Interfaces:**
- Consumes: same pipelinecheck functions as Task 2, plus
  `Config{IsEntrypoint, RequirePointFreeSeed}`.
- Produces: entrypoint gate over `RunCreateIssue`.

- [ ] **Step 1: Write the test**

```go
package github

import (
	"strings"
	"testing"

	pipelinecheck "github.com/dictyBase/fp-go-loom/pipelinecheck"
)

func TestPipelineContinuity(t *testing.T) {
	t.Parallel()

	pipelinecheck.Require(t, pipelinecheck.Config{
		Roots: []string{"."},
		IsEntrypoint: func(name string) bool {
			return name == "RunCreateIssue" || strings.HasPrefix(name, "run")
		},
		RequirePointFreeSeed:      true,
		RequirePointFreeBranching: true,
		RequireTrivialDelegate:    true,
	})
}
```

- [ ] **Step 2: Run it**

Run: `go test ./internal/app/github -run TestPipelineContinuity -v`
Expected: PASS, or fix flagged code in `github.go` per the message.

- [ ] **Step 3: Commit**

```sh
git add internal/app/github/pipelinecheck_test.go
git commit -m "test(app/github): add pipelinecheck mechanical gate"
```

### Task 4: Gate internal/app/webhook

**Files:**
- Create: `internal/app/webhook/pipelinecheck_test.go`

Same shape as Task 3 with the entrypoint name `RunOntoServer` and
package `webhook`.

- [ ] **Step 1: Write the test** — copy Task 3, rename package to
  `webhook`, match `RunOntoServer`.

- [ ] **Step 2: Run it**

Run: `go test ./internal/app/webhook -run TestPipelineContinuity -v`
Expected: PASS, or fix flagged code in `webhook.go` per the message.

- [ ] **Step 3: Commit**

```sh
git add internal/app/webhook/pipelinecheck_test.go
git commit -m "test(app/webhook): add pipelinecheck mechanical gate"
```

### Task 5: Upgrade the mailgun gate

**Files:**
- Modify: `internal/app/mailgun/pipelinecheck_test.go`

- [ ] **Step 1: Add `RequireTrivialDelegate: true` to the existing
  `pipelinecheck.Config` literal in `TestPipelineContinuity`.**

- [ ] **Step 2: Run it**

Run: `go test ./internal/app/mailgun -run TestPipelineContinuity -v`
Expected: PASS, or fix flagged code in `mailgun.go` per the message.

- [ ] **Step 3: Commit**

```sh
git add internal/app/mailgun/pipelinecheck_test.go
git commit -m "test(app/mailgun): enable trivial-delegate rule"
```

### Task 6: Prove the gate catches violations

One fixture, run once, delete it. This pins Review Focus items 1-5.

**Files:**
- Create then delete: `internal/datasource/gate_fixture.go` (NOT a
  `_test.go` file — `pipelinecheck.parseAll` scans only non-test
  `.go` files, so a `_test.go` fixture is invisible to the gate)

- [ ] **Step 1: Write a deliberate violation**

The trivial-delegate rule flags a wrapper whose entire body is
`return g(x)` forwarding its own parameter, referenced from an
`F.Pipe`/`F.Flow`/`IOE.*` site in the same package:

```go
package datasource

import (
	F "github.com/IBM/fp-go/v2/function"
)

func gateWrap(s string) string {
	return gateHelper(s)
}

func gatePipeline(s string) string {
	return F.Pipe1(s, gateWrap)
}

func gateHelper(s string) string { return s + "!" }
```

- [ ] **Step 2: Run and expect failure**

Run: `go test ./internal/datasource -run TestPipelineContinuity`
Expected: FAIL with `wrapper gateWrap only delegates to ...` at
`gate_fixture.go`. If the test still passes, the gate is broken —
stop and investigate before continuing.

- [ ] **Step 3: Delete the fixture**

Run: `rm internal/datasource/gate_fixture.go`

- [ ] **Step 4: Full gates**

```sh
go test ./...
golangci-lint fmt
golangci-lint run ./...
```

Expected: all green. Run `scripts/check-gopls.sh` when present;
otherwise `gopls check -severity=hint` on every tracked first-party
`.go` file, excluding `vendor/` and files whose first five lines
carry `// Code generated .* DO NOT EDIT.` (the mockery mocks and
`statik.go` qualify). Fix anything flagged before presenting.

- [ ] **Step 5: Commit any fixups and report**

```sh
git status --short   # must be clean or only untracked runtime files
```
