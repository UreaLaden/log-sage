package adapters

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
)

// Runner executes an external command and returns its stdout output.
type Runner func(name string, args ...string) ([]byte, error)

// KubectlAdapter fetches pod logs and pod description output via kubectl.
type KubectlAdapter struct {
	runner Runner
}

// New returns a KubectlAdapter that shells out to the real kubectl binary.
func New() *KubectlAdapter {
	return &KubectlAdapter{runner: defaultRunner}
}

// NewWithRunner returns a KubectlAdapter that uses r for command execution.
func NewWithRunner(r Runner) *KubectlAdapter {
	return &KubectlAdapter{runner: r}
}

func defaultRunner(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

// Fetch runs kubectl logs and kubectl describe pod for pod in namespace,
// concatenates their output, and returns it as an io.Reader.
func (a *KubectlAdapter) Fetch(pod, namespace string) (io.Reader, error) {
	logs, err := a.runner("kubectl", "logs", pod, "-n", namespace)
	if err != nil {
		return nil, fmt.Errorf("kubectl logs %s: %w", pod, err)
	}

	describe, err := a.runner("kubectl", "describe", "pod", pod, "-n", namespace)
	if err != nil {
		return nil, fmt.Errorf("kubectl describe pod %s: %w", pod, err)
	}

	var buf bytes.Buffer
	buf.Write(logs)
	buf.WriteByte('\n')
	buf.Write(describe)

	return &buf, nil
}
