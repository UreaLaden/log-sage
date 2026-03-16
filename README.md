# LogSage

LogSage analyzes log files and identifies the most likely root causes of system failures — returning ranked hypotheses with supporting evidence and suggested debugging steps.

<!-- TODO: add badges once CI and release pipeline are verified -->

---

<!-- TODO: replace placeholder demo with real example -->
## 10-Second Demo

LogSage scans raw logs and summarizes the most likely failure cause.

```bash
logsage ci --ci-summary build.log
```

Expected output:

```text
# placeholder: CI failure summary example output
Top Cause: <issue class> (<confidence> confidence)
Evidence:
- <evidence line 1>
- <evidence line 2>
Recommended Action:
- <recommended next step>
```

---

## Why LogSage

Debugging production failures means searching through thousands of log lines to find the one thing that went wrong.

LogSage speeds this up by:

1. Extracting meaningful signals from raw logs
2. Detecting known failure patterns across 12 issue classes
3. Ranking the most likely root causes by confidence
4. Returning evidence and recommended debugging steps

**Detected issue classes:** OutOfMemory, CrashLoopBackOff, ImagePullBackOff, ConnectionRefused, DNSFailure, TLSFailure, MissingEnvVar, PermissionDenied, DiskFull, DependencyTimeout, Panic, PortBindingFailure

---

## Installation

### Download binary

Prebuilt binaries for Linux, macOS, and Windows are available on the [releases](https://github.com/Urealaden/log-sage/releases) page.

> Maintainers: Release Please PRs use a PAT-backed `RELEASE_PLEASE_TOKEN` secret so normal PR workflows can run; `GITHUB_TOKEN` is not sufficient for that trigger path.

### Go install

```bash
go install github.com/Urealaden/log-sage/cmd/logsage@latest
```

### Build from source

```bash
git clone https://github.com/Urealaden/log-sage.git
cd log-sage
go build ./cmd/logsage
```

> Coming soon: Homebrew / Scoop / Chocolatey

---

## Quick Start

```bash
logsage ci build.log
logsage k8s pod my-pod --namespace production
```

---

## Usage

Analyze a log file:

```bash
logsage analyze server.log
```

Analyze piped input:

```bash
kubectl logs my-pod | logsage analyze --stdin
```

Analyze CI logs:

```bash
logsage ci build.log
```

Analyze a Kubernetes pod directly:

```bash
logsage k8s pod my-pod --namespace production
```

### Output flags

```bash
logsage analyze server.log --json       # machine-readable JSON
logsage analyze server.log --quiet      # class and confidence only
logsage analyze server.log --top 3      # limit to top N results
```

---

## CI Summary Mode

`--ci-summary` produces compact, human-readable output suitable for CI annotations and pull request comments.

```bash
logsage ci --ci-summary build.log
```

```text
# placeholder: CI failure summary example output
Top Cause: <issue class> (<confidence> confidence)
Evidence:
- <evidence line 1>
- <evidence line 2>
Recommended Action:
- <recommended next step>
```

`--ci-summary` is mutually exclusive with `--json` and `--quiet`.

---

## Kubernetes Logs

Fetch and analyze logs from a running Kubernetes pod:

```bash
logsage k8s pod my-pod --namespace production
```

LogSage fetches both `kubectl logs` and `kubectl describe pod` output and analyzes them as a single input.

---

## Example Output

Example output from the default analysis mode:

```text
# placeholder: full analysis output
Top Likely Causes

1. <issue class> (<confidence> confidence)

Evidence
- <matched log line>
- <matched log line>

Suggested Commands
- <kubectl or shell command>
- <kubectl or shell command>

Recommended Next Steps
- <step 1>
- <step 2>
```

---

## Architecture

LogSage uses a deterministic five-stage analysis pipeline:

```
normalize → extract → detect → score → recommend
```

- **normalize** — parse plaintext, JSON, or logfmt log entries into structured form
- **extract** — identify meaningful signals (errors, patterns, counts)
- **detect** — match signals against 12 known issue class definitions
- **score** — rank candidates by confidence (high / medium / low)
- **recommend** — generate suggested commands and next steps

The engine (`internal/engine`) is a reusable Go package with no dependency on the CLI. All analysis logic is accessible through a single `Analyze(ctx, DiagnosticInput) (*AnalysisResult, error)` interface.

---

## Testing

LogSage maintains an incident regression corpus at `testdata/incidents/`. Each incident directory contains a representative log sample and an `expected.json` schema defining the required issue classes, minimum confidence, and required evidence.

The corpus integration test (`internal/engine/corpus_test.go`) runs all samples against the engine on every build, preventing detection regressions.

To run the full test suite:

```bash
go test ./...
```

---

## Development

Build the binary:

```bash
go build ./cmd/logsage
```

Run all tests:

```bash
go test ./...
```

Run engine tests only:

```bash
go test ./internal/engine/...
```

Run the linter:

```bash
make run-lint
```

The linter uses `golangci-lint` with `errcheck`, `govet`, `staticcheck`, `unused`, `gofmt`, and `goimports`.

---

## Contributing

Contributions are welcome.

**Before opening a PR:**

1. Run `go test ./...` — all tests must pass
2. Run `make run-lint` — no lint errors
3. Add or update tests for any changed behavior
4. If adding a new failure detector, add a corresponding fixture under `testdata/incidents/`

**Project structure:**

```
cmd/logsage/           # CLI entry point
internal/
  engine/              # Public analysis interface
  normalization/       # Log parsing (plaintext, JSON, logfmt)
  extraction/          # Signal extraction
  detection/           # Issue class definitions and rule evaluation
  scoring/             # Hypothesis ranking and confidence
  recommendation/      # Next-step and command generation
pkg/types/             # Shared public types
testdata/incidents/    # Incident regression corpus
```

Areas where contributions are especially useful:

- New failure detector definitions (`internal/detection/`)
- Additional incident corpus samples (`testdata/incidents/`)
- Kubernetes adapter improvements (`internal/adapters/`)

---

## License

MIT
