package types

// SignalPattern defines a reusable extraction rule that identifies a specific
// signal in normalized log content.
type SignalPattern struct {
	Name            string
	SignalType      string
	MatchExpression string
}

// PatternMatch records one concrete occurrence of a SignalPattern in log
// content, including the matched text and deterministic source location.
type PatternMatch struct {
	PatternName string
	SignalType  string
	LineNumber  int
	Text        string
	Timestamp   string
}

// SignalSet contains the matches produced during extraction and serves as the
// shared input for later aggregation and detection stages.
type SignalSet struct {
	Matches []PatternMatch
}
