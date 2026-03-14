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