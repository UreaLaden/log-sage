# LogSage

🔎 Root-cause detection for logs.

LogSage analyzes logs and surfaces the **most likely causes of system failures** — with evidence and suggested debugging steps.

Instead of manually scanning thousands of lines, LogSage identifies common failure patterns like:

- CrashLoopBackOff
- Out-of-memory kills
- Connection refused
- DNS failures
- Missing environment variables


## Example

Input log:

```
panic: missing REDIS_URL
connection refused
CrashLoopBackOff restarting container
```

Run LogSage:
```
logsage analyze logs.txt
```

Output:
```
Top Likely Causes

1. Missing environment variable (high confidence)

Evidence

* panic: missing REDIS_URL
* configuration error during startup

Suggested Commands

* kubectl describe pod
* kubectl get configmap
```

LogSage turns raw logs into **actionable debugging insight**.

---

## Installation

### Go install

```
go install github.com/Urealaden/log-sage/cmd/logsage@latest
```

### Download binary

Prebuilt binaries for Linux, macOS, and Windows are available on the [releases](https://github.com/Urealaden/log-sage/releases) page.

> Package manager support (Homebrew, Scoop, Chocolatey) is planned for a post-v1 release.


## Usage

Analyze a log file:

```
logsage analyze logs.txt
```

Analyze piped logs:

```
kubectl logs my-pod | logsage analyze --stdin
```

Analyze a Kubernetes pod directly:
```
logsage k8s pod my-service --namespace production
```

Analyze CI logs:
```
logsage ci build.log
```

Output JSON:
```
logsage analyze logs.txt --json
```


## Detects Common Failures

LogSage detects:

- OutOfMemory / OOMKilled
- CrashLoopBackOff
- ConnectionRefused
- DNSFailure
- MissingEnvVar
- ImagePullBackOff
- TLSFailure
- PermissionDenied
- DiskFull
- DependencyTimeout
- Panic
- PortBindingFailure


## Why LogSage?

Debugging production failures often means manually searching through thousands of log lines.

LogSage speeds this up by:

1. Extracting meaningful signals from logs
2. Detecting known failure patterns
3. Ranking the most likely root causes
4. Showing evidence and recommended debugging steps

The goal is simple:

Turn logs into **diagnosis**, not just data.

---

## How It Works

LogSage uses a deterministic analysis engine.

```
              Logs
                ↓
          Normalization
                ↓
        Signal Extraction
                ↓
         Issue Detection
                ↓
        Hypothesis Ranking
                ↓
    Evidence + Recommendations
```

The engine is written in Go and designed to be reused by:

- CLI tools
- CI integrations
- Kubernetes debugging tools
- IDE plugins

---

## Contributing

LogSage is early and contributions are welcome.

Areas where help is especially useful:

- new failure detectors
- additional incident test cases
- performance improvements
- Kubernetes integrations


## Roadmap

- incident timeline reconstruction
- VS Code integration
- AI-assisted explanation improvements
- web UI / SaaS platform


## License

MIT
