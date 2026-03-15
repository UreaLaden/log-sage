# codex.md

Repository-specific execution notes for Codex.

## CLI verification rule

When verifying Go CLI behavior in this repository, do not run `go run` help/version checks in parallel.

Use this sequence instead:

1. Set writable Go cache locations:
   - `GOCACHE=C:\Users\leaun\AppData\Local\Temp\logsage-gocache`
   - `GOMODCACHE=C:\Users\leaun\AppData\Local\Temp\logsage-gomodcache`
2. Run `go test ./...` separately
3. Build the CLI once:
   - `go build -o C:\Users\leaun\AppData\Local\Temp\logsage-bin\logsage.exe ./cmd/logsage`
4. Verify the built binary sequentially:
   - `C:\Users\leaun\AppData\Local\Temp\logsage-bin\logsage.exe --help`
   - `C:\Users\leaun\AppData\Local\Temp\logsage-bin\logsage.exe version`
5. Apply short timeouts to each CLI verification command

Rationale: in this environment, parallel `go run` verification can appear to hang even when the CLI itself exits correctly. Building once and verifying the binary sequentially is more reliable.

## Repository Map Maintenance

When completing implementation work, Codex must evaluate whether the task changed repository structure or ownership.

Codex must update `.ai/repo-map.md` whenever a task:
- adds or removes directories or packages
- adds or changes CLI entrypoints
- adds or changes workflows or major config files
- changes package responsibilities
- changes the recommended navigation path for future agents

If no update is needed, Codex must explicitly state that in `.ai/handoffs/current.md`.

`.ai/repo-map.md` must reflect the current implemented repository state, not just the planned architecture.

## Handoff File Requirement

Before starting any implementation task, Codex must:

1. Read `.ai/handoffs/current.md` if it exists
2. Update `.ai/handoffs/current.md` before making code changes
3. Ensure `.ai/handoffs/current.md` tracks only the current task
4. Use `.ai/handoffs/template.md` as the required structure

At minimum, the pre-execution handoff update must include:
- task_id
- task_title
- status: in-progress
- owner: codex
- mapped backlog scope
- planned file changes

After completing work, Codex must update `.ai/handoffs/current.md` with:
- implementation steps
- files changed
- commands run
- test/build results
- test coverage
- assumptions made
- risks or concerns
- recommended next steps
- completion status

## Scope Control

Every implementation task must be tied to specific work items from `.ai/backlog/mapping.md`.

Codex must record in `.ai/handoffs/current.md`:
- Epic
- Feature
- Task(s)

Codex must stay within the mapped scope.

If out-of-scope work is required for compilation or task completion, Codex must record it under:

- Scope Variance
- Risks or Concerns
- Recommended Next Steps

Codex must not silently expand into sibling tasks, future tasks, or speculative cleanup.

## Test Coverage Requirement

When Codex adds or changes production code, it must evaluate whether new or updated tests are required.

Default rule:
- new logic requires new or updated tests

Minimum expectation:
- add at least one focused test covering the new behavior
- add an edge-case or failure-path test when applicable

Passing `go test ./...` alone is not sufficient if no task-specific tests were added for changed production code.

### Coverage Threshold

When Codex adds or changes production code, it must measure test coverage for the affected production package(s).

Minimum requirement:
- changed production package(s) must maintain at least 90% test coverage

Required coverage workflow:
1. Run `go test` for the affected package with coverage enabled
2. Generate a coverage profile
3. Record the result in `.ai/handoffs/current.md`

Example commands:
- `go test ./internal/normalize -coverprofile=coverage.out`
- `go tool cover -func=coverage.out`

Handoff requirement:
Codex must record in `.ai/handoffs/current.md`:
- test files added or updated
- coverage commands run
- coverage % for affected package(s)
- whether the 90% minimum was met

If coverage is below 90%, Codex must:
- add tests until the threshold is met, or
- stop and explicitly document why the threshold could not be met

Codex must include any new or modified test files in:
- Files Changed
- Test Coverage

This rule applies unless the task is explicitly marked test-exempt in the prompt or mapped backlog scope.

## Doc Comment Requirement

All exported Go symbols added or modified by Codex must have a doc comment.

Rules:
- Every exported function, type, method, and variable must have a comment beginning with its name
- Comments go immediately above the declaration with no blank line between them
- Unexported symbols do not require comments
- Do not add comments to unexported helpers unless the logic is non-obvious

Example:

```go
// ParseLogfmt reads logfmt-formatted lines from r and returns a slice of
// LogEntry values. Empty lines are skipped. ctx cancellation is respected
// between lines; if ctx is done the function returns immediately with ctx.Err().
func ParseLogfmt(ctx context.Context, r io.Reader) ([]LogEntry, error) {
```

Codex must verify doc comment presence before marking a task complete.
Missing doc comments on exported symbols are a required fix, not optional style.

## Verification Requirement

After implementation, Codex must run the verification commands required by the task prompt.

Codex must also run `make run-lint` at the end of every task as the final
verification step, unless the task is explicitly blocked before verification can
be completed.

At minimum, for normal Go implementation tasks:
- `go test ./...`
- `go build ./cmd/logsage`
- `make run-lint`

If additional package-level coverage commands are required, they must also be run and recorded in the handoff.

## Completion Output Rule

When Codex completes an implementation task in this repository, the final response must include a suggested commit message, even if the user did not explicitly ask for one.

The final response must also include short manual validation steps the user can run locally to spot-check the implemented behavior after automated verification succeeds.

All CodeRabbit review comments must be evaluated using the CodeRabbit
evaluation template before any changes are implemented.
