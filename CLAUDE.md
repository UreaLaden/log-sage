# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repository is

`log-sage` is the **AI scaffolding and documentation hub** for the LogSage project. It contains no application source code — only AI collaboration files tracked outside of the main source repository (`.ai/` and `.references/` are excluded from git via `.gitignore`).

LogSage is a **developer CLI tool and analysis engine that parses logs and identifies the most likely root causes of system failures.**

## Session initialization

At the start of every session, load context in this order:
1. `Agents.md` — project invariants and engineering guardrails (read first)
2. `.ai/agent-runtime.md` — agent roles, routing rules, handoff protocol
3. `.ai/repo-map.md` — directory layout and package responsibilities
4. `.ai/task.md` — epic/story status tracker

Load conditionally:
- `.ai/debug-playbook.md` — for debugging, failure analysis, or regression investigation
- `.references/tech-spec.md` — for architecture, feature behavior, or invariant decisions

After loading, confirm context is loaded and wait for the task.

## User shorthand

| What the user says | What it means |
|---|---|
| "review the handoff" / "review latest handoff" / "review current.md" | Read `.ai/handoffs/current.md` and evaluate the completed work |
| "evaluate the pr review" / "review the pr review" | Read `.ai/handoffs/coderabbit-eval.md` and evaluate its findings |

## CodeRabbit PR Review Evaluation

When the user asks to evaluate the PR review:

1. Enter plan mode
2. Read `.ai/handoffs/coderabbit-eval.md`
3. Locate each referenced file and verify whether the issue actually exists
4. Apply ACCEPT / PARTIAL / REJECT per the evaluation rules in the file
5. Output an implementation prompt for Codex — do NOT make code edits
6. Only make edits if the user explicitly requests them after seeing the plan

## Always reference current.md

Before answering any question about task status, what was completed, or what to work on next,
always read `.ai/handoffs/current.md` first. Do not rely on prior conversation context for
task state — the file is the source of truth.

## Handoff protocol

All handoff records must follow the structure defined in `.ai/handoffs/template.md`.
Copy that file to `.ai/handoffs/current.md` at the start of each new task.

`.ai/handoffs/current.md` is the single source of truth for task state. It enables Claude,
Codex, and ChatGPT to hand off work through repository files alone — no chat context required.

### Template reference

Every handoff uses `.ai/handoffs/template.md` as its structure. The template defines 11 sections
plus a YAML header. Do not invent an alternate format — copy the template and fill it in.

### Pre-execution requirement

Before modifying any file, the executing agent must:

1. Check whether `.ai/handoffs/current.md` exists and read it
2. If starting new work: copy `.ai/handoffs/template.md` to `.ai/handoffs/current.md`
3. Populate the YAML header: `task_id`, `task_title`, `owner`, `status: in-progress`, `started`, `last_updated`
4. Fill in **Section 1** (task summary — what problem this solves, what it touches, constraints)
5. Fill in **Section 2** (scope — specific backlog items covered; explicitly state what is out of scope)
6. Fill in **Section 3** (planned file changes — list every file expected to change before touching any)

This documents intent before any code or file is modified.

### Post-execution requirement

After completing work, update `current.md` with:

- **Section 4** — implementation steps taken (concrete, ordered)
- **Section 5** — files changed (list all modified or created files)
- **Section 6** — commands run (include results if relevant)
- **Section 7** — test/build results (pass/fail; explain any failures)
- **Section 8** — assumptions made during implementation
- **Section 9** — risks or concerns for the next agent
- **Section 10** — recommended next steps
- **Section 11** — final status: `completed` or `blocked` with reason

Also update the YAML header: set `status` to `completed` (or `blocked`) and `last_updated`.

### Agent interoperability

`current.md` enables handoffs between:

- **Claude** — planning, architecture, root cause analysis, spec interpretation
- **Codex** — focused implementation, bug fixes, tests
- **ChatGPT** — orchestration, cross-agent review

Any agent starting work must read `current.md` first if it exists. Never assume a prior agent's
intent from conversation context — derive it from the file.

### Execution rules

- Do not rely on conversation context for task continuity across sessions
- Repository files are the only persistent memory between agents and sessions
- `current.md` must always reflect the latest task state before any file changes are made
- Prepend new handoff entries to `current.md`; do not delete prior entries

## Backlog Generation

Whenever an ADO CSV import backlog is generated, Claude must:
1. **Write the CSV file** to `.references/tickets/` — never only print it in the response.
2. **Generate `.ai/backlog/mapping.md`** mirroring the Epic → Issue → Task hierarchy from the CSV.
3. **CSV columns must be:** `Work Item Type`, `Title`, `Description` only. Never include `Area Path`, `Iteration Path`, or `Parent` — `Area Path` and `Iteration Path` are org-specific and fail on import; `Parent` requires a numeric work item ID that does not exist pre-import. Hierarchy is established manually in ADO after import, using `mapping.md` as the reference.

The mapping must identify any orphaned or unresolved relationships.
The mapping file is required so backlog structure can be reviewed before import.
CSV files go under `.references/tickets/`; the mapping file always goes under `.ai/backlog/`.

## Build commands

| Command | Purpose |
|---|---|
| `go build ./cmd/logsage/...` | Build the CLI binary |
| `go test ./...` | Run all tests |
| `go test ./internal/engine/...` | Run engine tests only |
| `golangci-lint run` | Run linter |

Linter: `golangci-lint` with errcheck, govet, staticcheck, unused, gofmt, goimports.

## Architecture overview

LogSage is a Go CLI tool and reusable analysis engine. The CLI is a thin adapter; all analysis logic lives in the engine.

### Pipeline

```
Input Sources
   ↓
Normalization         — parse raw text into structured log entries
   ↓
Signal Extraction     — identify meaningful events (errors, patterns, counts)
   ↓
Issue Detection       — match signals against known issue class rules
   ↓
Hypothesis Scoring    — rank candidates by confidence
   ↓
Explanation Generation
   ↓
Recommendations
```

### Repository layout

```
cmd/logsage/           # CLI entry point
internal/
  adapters/            # Input source adapters (file, stdin, k8s, CI)
  engine/              # Public engine interface — Analyze(DiagnosticInput) -> AnalysisResult
  normalization/       # Log parsing (plaintext, JSON, logfmt)
  extraction/          # Signal extraction from normalized entries
  detection/           # Issue class definitions and rule evaluation
  scoring/             # Hypothesis ranking and confidence calculation
  recommendation/      # Next-step and command generation
pkg/types/             # Shared public types
testdata/incidents/    # Real/synthetic incident log corpus
```

## CLI Boundary Rule

The CLI must depend only on `internal/engine`.

Allowed flow:
cmd/logsage -> internal/engine -> normalize/extraction/detection/scoring/recommendation

Disallowed:
- cmd/logsage importing internal/detection directly
- cmd/logsage importing internal/scoring directly
- cmd/logsage importing internal/recommendation directly
- cmd/logsage importing internal/extraction directly
- CLI flags bypassing engine orchestration

The engine is the single orchestration boundary for analysis behavior.
All user-facing rendering, formatting, and output-mode handling belong in the CLI.
All analysis, scoring, and recommendation logic belong behind `internal/engine`.

### Engine API

```go
// Engine interface
engine.Analyze(ctx, DiagnosticInput) -> (*AnalysisResult, error)

// Key types
DiagnosticInput  { Logs []string, SourceType, Metadata }
AnalysisResult   { Summary, TopCauses []Hypothesis, Evidence, Unknowns, RecommendedCommands, RecommendedNextSteps }
Hypothesis       { IssueClass, Confidence (high|medium|low), Score, Evidence, Explanation }
```

### Issue class taxonomy (MVP — 12 classes)

`OutOfMemory`, `CrashLoopBackOff`, `ImagePullBackOff`, `ConnectionRefused`, `DNSFailure`, `TLSFailure`, `MissingEnvVar`, `PermissionDenied`, `DiskFull`, `DependencyTimeout`, `Panic`, `PortBindingFailure`

Each class defines: primary signals, corroborating signals, confidence rules, explanation template, next steps.

### CLI commands

```
logsage analyze <file>
logsage analyze --stdin
logsage k8s pod <pod-name>
logsage ci <log-file>
```

Flags: `--json`, `--quiet`, `--top N`, `--namespace`

### Key invariants

- The engine never asserts a hypothesis without at least one signal match in the input logs.
- Confidence levels are derived from defined rules — not hardcoded.
- Every `Hypothesis` must reference at least one `Evidence` entry.
- `internal/engine` has no dependency on the CLI package.
- The CLI calls only `internal/engine` for analysis — never detection/scoring/extraction directly.

## Task routing

| Task type | Preferred agent |
|---|---|
| Spec interpretation, architecture, root cause analysis | Claude |
| Bug fix with known cause, small implementation, tests | Codex |
| Cross-cutting design change, ambiguous behavior | Claude → Codex |

**High-risk files — always use Claude → Codex:**
- `internal/engine/engine.go` — public engine interface and invariants
- `internal/detection/` — issue class taxonomy rules
- `internal/scoring/` — confidence calculation logic
- `pkg/types/` — shared public types (breaking changes ripple everywhere)

**Low-risk — safe for direct Codex implementation:**
- Individual issue class definitions, CLI flag wiring, test corpus entries, `.ai/` files

## Technology

- **Language:** Go 1.24+
- **CLI framework:** `cobra`
- **No external runtime dependencies** — single static binary
- **Testing:** standard `testing` package + table-driven tests against `testdata/incidents/` corpus

## Spec

Full MVP technical specification: `.references/tech-spec.md`

## Required Task Closeout

When a mapped task has been implemented, verified, and is ready to close, Claude must perform the following closeout steps before finishing:

1. Read `.ai/handoffs/current.md`
2. Read `.ai/handoffs/pr-template.md`
3. Generate a completed PR summary with no placeholders
4. Write the result to `.ai/handoffs/pr-summary.md`

This is required whenever:
- implementation is complete
- verification commands have passed
- required test coverage has been recorded
- the task is ready to close

The PR summary must be grounded only in the active handoff and repository artifacts. Do not invent scope, verification, or coverage details.