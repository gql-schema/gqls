package gql

import "github.com/antlr4-go/antlr/v4"

// ErrorListener is a custom error listener for ANTLR that collects syntax errors.
type ErrorListener struct {
	// SyntaxErrors contains a list of syntax errors encountered during parsing.
	SyntaxErrors []string

	*antlr.DefaultErrorListener
}

var _ antlr.ErrorListener = (*ErrorListener)(nil)

// SyntaxError is called when a syntax error is encountered.
func (l *ErrorListener) SyntaxError(_ antlr.Recognizer, _ any, _, _ int, msg string, _ antlr.RecognitionException) {
	l.SyntaxErrors = append(l.SyntaxErrors, msg)
}
