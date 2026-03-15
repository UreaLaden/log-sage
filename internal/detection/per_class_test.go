package detection

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Urealaden/log-sage-temp/internal/extraction"
	"github.com/Urealaden/log-sage-temp/internal/normalize"
	"github.com/Urealaden/log-sage-temp/pkg/types"
)

func classFromRegistry(t *testing.T, name string) types.IssueClass {
	t.Helper()

	for _, class := range IssueRegistry {
		if class.Name == name {
			return class
		}
	}

	t.Fatalf("issue class %q not found in registry", name)
	return types.IssueClass{}
}

func signalsFromFixture(t *testing.T, fixture string, patterns []types.SignalPattern) types.SignalSet {
	t.Helper()

	path := filepath.Join("..", "..", "testdata", "logs", fixture)
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open fixture %q: %v", fixture, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("close fixture %q: %v", fixture, err)
		}
	}()

	lines, err := normalize.Normalize(context.Background(), file)
	if err != nil {
		t.Fatalf("normalize fixture %q: %v", fixture, err)
	}

	return extraction.ExtractSignals(lines, patterns)
}

func assertOneHypothesis(t *testing.T, got []types.Hypothesis, wantClass string) {
	t.Helper()

	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}

	if got[0].IssueClass != wantClass {
		t.Fatalf("IssueClass = %q, want %q", got[0].IssueClass, wantClass)
	}

	if len(got[0].Evidence) == 0 {
		t.Fatal("expected evidence to be present")
	}
}

func assertNoHypothesis(t *testing.T, got []types.Hypothesis) {
	t.Helper()

	if len(got) != 0 {
		t.Fatalf("got %d hypotheses, want 0", len(got))
	}
}

func assertDeterministicDetection(t *testing.T, class types.IssueClass, signals types.SignalSet) {
	t.Helper()

	detector := DefaultDetector{Class: class}
	first := detector.Detect(signals)
	second := detector.Detect(signals)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("Detect() results differ across repeated calls: %#v != %#v", first, second)
	}
}

func assertEmptySignalsSafe(t *testing.T, class types.IssueClass) {
	t.Helper()

	detector := DefaultDetector{Class: class}

	assertNoHypothesis(t, detector.Detect(types.SignalSet{}))
	assertNoHypothesis(t, detector.Detect(types.SignalSet{Matches: nil}))
	assertNoHypothesis(t, detector.Detect(types.SignalSet{Matches: []types.PatternMatch{}}))
}

func patternsForClass(class types.IssueClass) []types.SignalPattern {
	return append(append([]types.SignalPattern{}, class.PrimarySignals...), class.CorroboratingSignals...)
}

func TestOutOfMemoryDetector(t *testing.T) {
	t.Parallel()
	class := classFromRegistry(t, "OutOfMemory")

	t.Run("oom-sample.log triggers hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "oom-sample.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertOneHypothesis(t, got, "OutOfMemory")
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("healthy-startup.log produces no hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "healthy-startup.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertNoHypothesis(t, got)
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("empty signals produce no hypothesis", func(t *testing.T) {
		t.Parallel()
		assertEmptySignalsSafe(t, class)
	})
}

func TestCrashLoopBackOffDetector(t *testing.T) {
	t.Parallel()
	class := classFromRegistry(t, "CrashLoopBackOff")

	t.Run("k8s-pod-sample.log triggers hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "k8s-pod-sample.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertOneHypothesis(t, got, "CrashLoopBackOff")
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("healthy-startup.log produces no hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "healthy-startup.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertNoHypothesis(t, got)
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("empty signals produce no hypothesis", func(t *testing.T) {
		t.Parallel()
		assertEmptySignalsSafe(t, class)
	})
}

func TestImagePullBackOffDetector(t *testing.T) {
	t.Parallel()
	class := classFromRegistry(t, "ImagePullBackOff")

	t.Run("imagepull-sample.log triggers hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "imagepull-sample.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertOneHypothesis(t, got, "ImagePullBackOff")
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("healthy-startup.log produces no hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "healthy-startup.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertNoHypothesis(t, got)
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("empty signals produce no hypothesis", func(t *testing.T) {
		t.Parallel()
		assertEmptySignalsSafe(t, class)
	})
}

func TestConnectionRefusedDetector(t *testing.T) {
	t.Parallel()
	class := classFromRegistry(t, "ConnectionRefused")

	t.Run("k8s-pod-sample.log triggers hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "k8s-pod-sample.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertOneHypothesis(t, got, "ConnectionRefused")
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("healthy-startup.log produces no hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "healthy-startup.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertNoHypothesis(t, got)
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("empty signals produce no hypothesis", func(t *testing.T) {
		t.Parallel()
		assertEmptySignalsSafe(t, class)
	})
}

func TestDNSFailureDetector(t *testing.T) {
	t.Parallel()
	class := classFromRegistry(t, "DNSFailure")

	t.Run("dns-sample.log triggers hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "dns-sample.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertOneHypothesis(t, got, "DNSFailure")
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("healthy-startup.log produces no hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "healthy-startup.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertNoHypothesis(t, got)
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("empty signals produce no hypothesis", func(t *testing.T) {
		t.Parallel()
		assertEmptySignalsSafe(t, class)
	})
}

func TestTLSFailureDetector(t *testing.T) {
	t.Parallel()
	class := classFromRegistry(t, "TLSFailure")

	t.Run("tls-sample.log triggers hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "tls-sample.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertOneHypothesis(t, got, "TLSFailure")
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("healthy-startup.log produces no hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "healthy-startup.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertNoHypothesis(t, got)
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("empty signals produce no hypothesis", func(t *testing.T) {
		t.Parallel()
		assertEmptySignalsSafe(t, class)
	})
}

func TestMissingEnvVarDetector(t *testing.T) {
	t.Parallel()
	class := classFromRegistry(t, "MissingEnvVar")

	t.Run("missingenv-sample.log triggers hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "missingenv-sample.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertOneHypothesis(t, got, "MissingEnvVar")
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("healthy-startup.log produces no hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "healthy-startup.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertNoHypothesis(t, got)
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("empty signals produce no hypothesis", func(t *testing.T) {
		t.Parallel()
		assertEmptySignalsSafe(t, class)
	})
}

func TestPermissionDeniedDetector(t *testing.T) {
	t.Parallel()
	class := classFromRegistry(t, "PermissionDenied")

	t.Run("permission-sample.log triggers hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "permission-sample.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertOneHypothesis(t, got, "PermissionDenied")
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("healthy-startup.log produces no hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "healthy-startup.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertNoHypothesis(t, got)
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("empty signals produce no hypothesis", func(t *testing.T) {
		t.Parallel()
		assertEmptySignalsSafe(t, class)
	})
}

func TestDiskFullDetector(t *testing.T) {
	t.Parallel()
	class := classFromRegistry(t, "DiskFull")

	t.Run("diskfull-sample.log triggers hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "diskfull-sample.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertOneHypothesis(t, got, "DiskFull")
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("healthy-startup.log produces no hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "healthy-startup.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertNoHypothesis(t, got)
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("empty signals produce no hypothesis", func(t *testing.T) {
		t.Parallel()
		assertEmptySignalsSafe(t, class)
	})
}

func TestDependencyTimeoutDetector(t *testing.T) {
	t.Parallel()
	class := classFromRegistry(t, "DependencyTimeout")

	t.Run("ci-pipeline-sample.log triggers hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "ci-pipeline-sample.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertOneHypothesis(t, got, "DependencyTimeout")
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("healthy-startup.log produces no hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "healthy-startup.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertNoHypothesis(t, got)
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("empty signals produce no hypothesis", func(t *testing.T) {
		t.Parallel()
		assertEmptySignalsSafe(t, class)
	})
}

func TestPanicDetector(t *testing.T) {
	t.Parallel()
	class := classFromRegistry(t, "Panic")

	t.Run("stacktrace-sample.log triggers hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "stacktrace-sample.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertOneHypothesis(t, got, "Panic")
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("healthy-startup.log produces no hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "healthy-startup.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertNoHypothesis(t, got)
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("empty signals produce no hypothesis", func(t *testing.T) {
		t.Parallel()
		assertEmptySignalsSafe(t, class)
	})
}

func TestPortBindingFailureDetector(t *testing.T) {
	t.Parallel()
	class := classFromRegistry(t, "PortBindingFailure")

	t.Run("portbind-sample.log triggers hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "portbind-sample.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertOneHypothesis(t, got, "PortBindingFailure")
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("healthy-startup.log produces no hypothesis", func(t *testing.T) {
		t.Parallel()
		signals := signalsFromFixture(t, "healthy-startup.log", patternsForClass(class))
		got := (DefaultDetector{Class: class}).Detect(signals)
		assertNoHypothesis(t, got)
		assertDeterministicDetection(t, class, signals)
	})

	t.Run("empty signals produce no hypothesis", func(t *testing.T) {
		t.Parallel()
		assertEmptySignalsSafe(t, class)
	})
}
