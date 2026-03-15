package extraction

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Urealaden/log-sage-temp/internal/normalize"
	"github.com/Urealaden/log-sage-temp/pkg/types"
)

func TestExtractSignalsWithFixtures(t *testing.T) {
	t.Parallel()

	dnsFailurePatterns := []types.SignalPattern{
		{Name: "dns-no-such-host", SignalType: "DNSFailure", MatchExpression: "no such host"},
		{Name: "dns-name-resolution", SignalType: "DNSFailure", MatchExpression: "name resolution"},
	}

	tlsFailurePatterns := []types.SignalPattern{
		{Name: "tls-unknown-authority", SignalType: "TLSFailure", MatchExpression: "certificate signed by unknown authority"},
		{Name: "tls-handshake", SignalType: "TLSFailure", MatchExpression: "tls: handshake failure"},
	}

	panicPatterns := []types.SignalPattern{
		{Name: "panic-keyword", SignalType: "Panic", MatchExpression: "panic:"},
		{Name: "goroutine-dump", SignalType: "Panic", MatchExpression: "goroutine"},
	}

	connectionRefusedPatterns := []types.SignalPattern{
		{Name: "conn-refused", SignalType: "ConnectionRefused", MatchExpression: "connection refused"},
	}

	portBindingFailurePatterns := []types.SignalPattern{
		{Name: "bind-address-in-use", SignalType: "PortBindingFailure", MatchExpression: "address already in use"},
		{Name: "bind-prefix", SignalType: "PortBindingFailure", MatchExpression: "bind: "},
	}

	tests := []struct {
		name                   string
		fixture                string
		patterns               []types.SignalPattern
		expectedSignalType     string
		minimumExpectedMatches int
		wantZeroMatches        bool
	}{
		{
			name:                   "dns fixture produces dns failure matches",
			fixture:                "dns-sample.log",
			patterns:               dnsFailurePatterns,
			expectedSignalType:     "DNSFailure",
			minimumExpectedMatches: 2,
		},
		{
			name:            "healthy fixture produces zero matches for dns patterns",
			fixture:         "healthy-startup.log",
			patterns:        dnsFailurePatterns,
			wantZeroMatches: true,
		},
		{
			name:                   "tls fixture produces tls failure matches",
			fixture:                "tls-sample.log",
			patterns:               tlsFailurePatterns,
			expectedSignalType:     "TLSFailure",
			minimumExpectedMatches: 2,
		},
		{
			name:            "healthy fixture produces zero matches for tls patterns",
			fixture:         "healthy-startup.log",
			patterns:        tlsFailurePatterns,
			wantZeroMatches: true,
		},
		{
			name:                   "stacktrace fixture produces panic matches",
			fixture:                "stacktrace-sample.log",
			patterns:               panicPatterns,
			expectedSignalType:     "Panic",
			minimumExpectedMatches: 2,
		},
		{
			name:            "healthy fixture produces zero matches for panic patterns",
			fixture:         "healthy-startup.log",
			patterns:        panicPatterns,
			wantZeroMatches: true,
		},
		{
			name:                   "k8s pod fixture produces connection refused matches",
			fixture:                "k8s-pod-sample.log",
			patterns:               connectionRefusedPatterns,
			expectedSignalType:     "ConnectionRefused",
			minimumExpectedMatches: 2,
		},
		{
			name:            "healthy fixture produces zero matches for connection refused patterns",
			fixture:         "healthy-startup.log",
			patterns:        connectionRefusedPatterns,
			wantZeroMatches: true,
		},
		{
			name:                   "port bind fixture produces port binding failure matches",
			fixture:                "portbind-sample.log",
			patterns:               portBindingFailurePatterns,
			expectedSignalType:     "PortBindingFailure",
			minimumExpectedMatches: 2,
		},
		{
			name:            "healthy fixture produces zero matches for port binding patterns",
			fixture:         "healthy-startup.log",
			patterns:        portBindingFailurePatterns,
			wantZeroMatches: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			entries := loadFixtureEntries(t, tt.fixture)
			got := ExtractSignals(entries, tt.patterns)
			gotAgain := ExtractSignals(entries, tt.patterns)

			if tt.wantZeroMatches {
				if len(got.Matches) != 0 {
					t.Fatalf("match count = %d, want 0", len(got.Matches))
				}
				return
			}

			if len(got.Matches) < tt.minimumExpectedMatches {
				t.Fatalf("match count = %d, want at least %d", len(got.Matches), tt.minimumExpectedMatches)
			}

			for i, match := range got.Matches {
				if match.SignalType != tt.expectedSignalType {
					t.Fatalf("match[%d].SignalType = %q, want %q", i, match.SignalType, tt.expectedSignalType)
				}
			}

			assertStableMatchOrder(t, got.Matches, gotAgain.Matches)
		})
	}
}

func loadFixtureEntries(t *testing.T, fixture string) []normalize.Line {
	t.Helper()

	path := filepath.Join("..", "..", "testdata", "logs", fixture)
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open fixture %q: %v", fixture, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("close file: %v", err)
		}
	}()

	entries, err := normalize.Normalize(context.Background(), file)
	if err != nil {
		t.Fatalf("normalize fixture %q: %v", fixture, err)
	}

	return entries
}

func assertStableMatchOrder(t *testing.T, matches []types.PatternMatch, matchesAgain []types.PatternMatch) {
	t.Helper()

	if !reflect.DeepEqual(matches, matchesAgain) {
		t.Fatalf("matches are not stable across repeated runs: %#v != %#v", matches, matchesAgain)
	}

	previousLineNumber := 0

	for i, match := range matches {
		if match.LineNumber < previousLineNumber {
			t.Fatalf("match[%d].LineNumber = %d, previous line number = %d", i, match.LineNumber, previousLineNumber)
		}
		previousLineNumber = match.LineNumber
	}
}
