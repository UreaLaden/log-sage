package detection

import "github.com/Urealaden/log-sage-temp/pkg/types"

// IssueRegistry contains the canonical MVP issue-class definitions consumed by
// later detection infrastructure.
var IssueRegistry = []types.IssueClass{
	{
		Name: "OutOfMemory",
		PrimarySignals: []types.SignalPattern{
			{Name: "oom-killed", SignalType: "OutOfMemory", MatchExpression: "OOMKilled", Weight: 1.0},
			{Name: "out-of-memory", SignalType: "OutOfMemory", MatchExpression: "out of memory", Weight: 0.9},
			{Name: "cannot-allocate-memory", SignalType: "OutOfMemory", MatchExpression: "cannot allocate memory", Weight: 0.9},
			{Name: "exit-code-137", SignalType: "OutOfMemory", MatchExpression: "exit code 137", Weight: 0.8},
		},
		CorroboratingSignals: []types.SignalPattern{
			{Name: "memory-allocation-error", SignalType: "OutOfMemory", MatchExpression: "allocation", Weight: 0.3},
		},
		ExplanationTemplate: "The container exceeded its memory limit and was terminated by the runtime.",
		NextSteps: []string{
			"Inspect pod restart history and termination reason.",
			"Compare container memory limits with observed workload usage.",
			"Review recent deploy or traffic changes that may have increased memory pressure.",
		},
		Commands: []string{
			"kubectl describe pod <pod>",
			"kubectl top pod <pod>",
		},
	},
	{
		Name: "CrashLoopBackOff",
		PrimarySignals: []types.SignalPattern{
			{Name: "crashloopbackoff", SignalType: "CrashLoopBackOff", MatchExpression: "CrashLoopBackOff", Weight: 1.0},
			{Name: "back-off-restarting", SignalType: "CrashLoopBackOff", MatchExpression: "Back-off restarting failed container", Weight: 0.9},
			{Name: "restarting-container", SignalType: "CrashLoopBackOff", MatchExpression: "restarting failed container", Weight: 0.8},
		},
		CorroboratingSignals: []types.SignalPattern{
			{Name: "panic", SignalType: "CrashLoopBackOff", MatchExpression: "panic:", Weight: 0.4},
			{Name: "fatal", SignalType: "CrashLoopBackOff", MatchExpression: "fatal", Weight: 0.3},
		},
		ExplanationTemplate: "The container is repeatedly crashing shortly after startup.",
		NextSteps: []string{
			"Review previous container logs to find the first fatal error.",
			"Check startup configuration and dependency availability.",
			"Confirm the container can complete initialization before probes fail.",
		},
		Commands: []string{
			"kubectl logs <pod> --previous",
			"kubectl describe pod <pod>",
		},
	},
	{
		Name: "ImagePullBackOff",
		PrimarySignals: []types.SignalPattern{
			{Name: "imagepullbackoff", SignalType: "ImagePullBackOff", MatchExpression: "ImagePullBackOff", Weight: 1.0},
			{Name: "errimagepull", SignalType: "ImagePullBackOff", MatchExpression: "ErrImagePull", Weight: 0.9},
			{Name: "pull-access-denied", SignalType: "ImagePullBackOff", MatchExpression: "pull access denied", Weight: 0.9},
		},
		CorroboratingSignals: []types.SignalPattern{
			{Name: "manifest-unknown", SignalType: "ImagePullBackOff", MatchExpression: "manifest unknown", Weight: 0.3},
		},
		ExplanationTemplate: "The container image could not be pulled successfully.",
		NextSteps: []string{
			"Verify the image name and tag are correct.",
			"Check registry credentials and image pull secrets.",
			"Confirm the image exists and is accessible from the cluster.",
		},
		Commands: []string{
			"kubectl describe pod <pod>",
			"kubectl get secret",
		},
	},
	{
		Name: "ConnectionRefused",
		PrimarySignals: []types.SignalPattern{
			{Name: "connection-refused", SignalType: "ConnectionRefused", MatchExpression: "connection refused", Weight: 1.0},
			{Name: "econnrefused", SignalType: "ConnectionRefused", MatchExpression: "ECONNREFUSED", Weight: 0.9},
			{Name: "dial-tcp", SignalType: "ConnectionRefused", MatchExpression: "dial tcp", Weight: 0.7},
			{Name: "connect-connection-refused", SignalType: "ConnectionRefused", MatchExpression: "connect: connection refused", Weight: 0.9},
		},
		CorroboratingSignals: []types.SignalPattern{
			{Name: "retrying-connection", SignalType: "ConnectionRefused", MatchExpression: "retrying connection", Weight: 0.3},
		},
		ExplanationTemplate: "The application attempted to connect to a service that refused the connection.",
		NextSteps: []string{
			"Confirm the dependency service is running and listening on the expected port.",
			"Check service discovery and target endpoint configuration.",
			"Review whether the dependency is failing its own startup checks.",
		},
		Commands: []string{
			"kubectl get svc",
			"kubectl describe svc <service>",
		},
	},
	{
		Name: "DNSFailure",
		PrimarySignals: []types.SignalPattern{
			{Name: "no-such-host", SignalType: "DNSFailure", MatchExpression: "no such host", Weight: 1.0},
			{Name: "nxdomain", SignalType: "DNSFailure", MatchExpression: "NXDOMAIN", Weight: 0.9},
			{Name: "lookup-failed", SignalType: "DNSFailure", MatchExpression: "lookup failed", Weight: 0.8},
			{Name: "name-resolution", SignalType: "DNSFailure", MatchExpression: "temporary failure in name resolution", Weight: 0.9},
		},
		CorroboratingSignals: []types.SignalPattern{
			{Name: "retrying-dns", SignalType: "DNSFailure", MatchExpression: "retrying DNS lookup", Weight: 0.3},
		},
		ExplanationTemplate: "The application could not resolve the hostname of a dependency.",
		NextSteps: []string{
			"Verify the target service name and namespace are correct.",
			"Check cluster DNS health and pod DNS configuration.",
			"Confirm the dependency service exists and is discoverable.",
		},
		Commands: []string{
			"kubectl exec -it <pod> -- nslookup <service>",
			"kubectl get svc",
		},
	},
	{
		Name: "TLSFailure",
		PrimarySignals: []types.SignalPattern{
			{Name: "unknown-authority", SignalType: "TLSFailure", MatchExpression: "certificate signed by unknown authority", Weight: 1.0},
			{Name: "x509", SignalType: "TLSFailure", MatchExpression: "x509:", Weight: 0.9},
			{Name: "tls-prefix", SignalType: "TLSFailure", MatchExpression: "tls:", Weight: 0.8},
		},
		CorroboratingSignals: []types.SignalPattern{
			{Name: "handshake-failure", SignalType: "TLSFailure", MatchExpression: "handshake failure", Weight: 0.4},
		},
		ExplanationTemplate: "The application failed to establish a trusted TLS connection.",
		NextSteps: []string{
			"Verify the server certificate chain and trusted CA bundle.",
			"Check hostname matching and certificate expiration.",
			"Confirm the client is using the expected TLS settings.",
		},
		Commands: []string{
			"openssl s_client -connect <host>:<port>",
			"kubectl describe secret <tls-secret>",
		},
	},
	{
		Name: "MissingEnvVar",
		PrimarySignals: []types.SignalPattern{
			{Name: "missing-environment-variable", SignalType: "MissingEnvVar", MatchExpression: "missing environment variable", Weight: 1.0},
			{Name: "required-env", SignalType: "MissingEnvVar", MatchExpression: "required env", Weight: 0.9},
			{Name: "panic-getenv", SignalType: "MissingEnvVar", MatchExpression: "panic: getenv", Weight: 0.9},
			{Name: "config-not-found", SignalType: "MissingEnvVar", MatchExpression: "config not found", Weight: 0.8},
		},
		CorroboratingSignals: []types.SignalPattern{
			{Name: "startup-error", SignalType: "MissingEnvVar", MatchExpression: "startup error", Weight: 0.3},
			{Name: "configuration-error", SignalType: "MissingEnvVar", MatchExpression: "configuration", Weight: 0.3},
		},
		ExplanationTemplate: "The application failed to start due to missing configuration.",
		NextSteps: []string{
			"Check required environment variables and mounted configuration.",
			"Verify referenced ConfigMaps and Secrets exist and are populated.",
			"Review the startup path for missing configuration defaults.",
		},
		Commands: []string{
			"kubectl describe pod <pod>",
			"kubectl get configmap",
			"kubectl get secret",
		},
	},
	{
		Name: "PermissionDenied",
		PrimarySignals: []types.SignalPattern{
			{Name: "permission-denied", SignalType: "PermissionDenied", MatchExpression: "permission denied", Weight: 1.0},
			{Name: "eacces", SignalType: "PermissionDenied", MatchExpression: "EACCES", Weight: 0.9},
			{Name: "forbidden-403", SignalType: "PermissionDenied", MatchExpression: "403 Forbidden", Weight: 0.8},
		},
		ExplanationTemplate: "The application was blocked by missing file, process, or API permissions.",
		NextSteps: []string{
			"Identify which path, resource, or API request was denied.",
			"Verify service account, role bindings, and file ownership.",
			"Check whether the runtime user has the expected access.",
		},
		Commands: []string{
			"kubectl auth can-i --list --as system:serviceaccount:<ns>:<sa>",
			"kubectl describe pod <pod>",
		},
	},
	{
		Name: "DiskFull",
		PrimarySignals: []types.SignalPattern{
			{Name: "no-space-left", SignalType: "DiskFull", MatchExpression: "no space left on device", Weight: 1.0},
			{Name: "enospc", SignalType: "DiskFull", MatchExpression: "ENOSPC", Weight: 0.9},
			{Name: "disk-quota-exceeded", SignalType: "DiskFull", MatchExpression: "disk quota exceeded", Weight: 0.9},
		},
		ExplanationTemplate: "The workload ran out of writable disk capacity.",
		NextSteps: []string{
			"Determine which filesystem or volume is full.",
			"Review recent log growth, cache growth, or artifact writes.",
			"Decide whether to clean up data or raise storage limits.",
		},
		Commands: []string{
			"kubectl describe pod <pod>",
			"df -h",
		},
	},
	{
		Name: "DependencyTimeout",
		PrimarySignals: []types.SignalPattern{
			{Name: "context-deadline-exceeded", SignalType: "DependencyTimeout", MatchExpression: "context deadline exceeded", Weight: 1.0},
			{Name: "timeout-after", SignalType: "DependencyTimeout", MatchExpression: "timeout after", Weight: 0.9},
			{Name: "network-timeout", SignalType: "DependencyTimeout", MatchExpression: "network timeout", Weight: 0.8},
			{Name: "io-timeout", SignalType: "DependencyTimeout", MatchExpression: "i/o timeout", Weight: 0.8},
		},
		CorroboratingSignals: []types.SignalPattern{
			{Name: "retrying-step", SignalType: "DependencyTimeout", MatchExpression: "retrying step", Weight: 0.3},
		},
		ExplanationTemplate: "The application timed out while waiting on an external dependency.",
		NextSteps: []string{
			"Identify which dependency call timed out.",
			"Check network reachability and dependency health.",
			"Review timeout thresholds and retry settings for that operation.",
		},
		Commands: []string{
			"kubectl logs <pod>",
			"kubectl get endpoints <service>",
		},
	},
	{
		Name: "Panic",
		PrimarySignals: []types.SignalPattern{
			{Name: "panic-prefix", SignalType: "Panic", MatchExpression: "panic:", Weight: 1.0},
			{Name: "goroutine", SignalType: "Panic", MatchExpression: "goroutine", Weight: 0.8},
			{Name: "runtime-error", SignalType: "Panic", MatchExpression: "runtime error:", Weight: 0.9},
		},
		CorroboratingSignals: []types.SignalPattern{
			{Name: "stack-frame", SignalType: "Panic", MatchExpression: ".go:", Weight: 0.3},
		},
		ExplanationTemplate: "The application encountered an unrecovered panic and aborted execution.",
		NextSteps: []string{
			"Inspect the first panic line and the top stack frames.",
			"Identify the code path and input that triggered the panic.",
			"Check whether a missing dependency or configuration error caused the failure.",
		},
		Commands: []string{
			"kubectl logs <pod> --previous",
			"kubectl describe pod <pod>",
		},
	},
	{
		Name: "PortBindingFailure",
		PrimarySignals: []types.SignalPattern{
			{Name: "address-already-in-use", SignalType: "PortBindingFailure", MatchExpression: "address already in use", Weight: 1.0},
			{Name: "eaddrinuse", SignalType: "PortBindingFailure", MatchExpression: "EADDRINUSE", Weight: 0.9},
			{Name: "bind-prefix", SignalType: "PortBindingFailure", MatchExpression: "bind: ", Weight: 0.8},
		},
		ExplanationTemplate: "The application failed to bind the configured listening port.",
		NextSteps: []string{
			"Confirm the application is configured to use the expected port.",
			"Check whether another process is already bound to that port.",
			"Review recent config changes that may have duplicated listeners.",
		},
		Commands: []string{
			"kubectl logs <pod>",
			"ss -ltnp",
		},
	},
}
