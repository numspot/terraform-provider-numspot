package lib

import (
	"context"
	"fmt"
	"strings"

	opl "github.com/ory/keto/proto/ory/keto/opl/v1alpha1"
)

// SourcePosition of the [ParseError].
type SourcePosition struct {
	Line   uint32
	Column uint32
}

// ParseError returned in case [syntaxClient.Check] call returned a parsing error.
// This exposes [github.com/ory/keto/proto/ory/keto/opl/v1alpha1.ParseError].
type ParseError struct {
	Message       string
	StartPosition *SourcePosition
	EndPosition   *SourcePosition
}

// Error formats in the 'message: start-end' format.
func (err ParseError) Error() string {
	builder := new(strings.Builder)
	fmt.Fprintf(builder, "%s:", err.Message)
	if err.StartPosition != nil {
		fmt.Fprintf(builder, " %d:%d", err.StartPosition.Line, err.StartPosition.Column)
	}
	if err.EndPosition != nil {
		fmt.Fprintf(builder, "-%d:%d", err.EndPosition.Line, err.EndPosition.Column)
	}
	return builder.String()
}

// newParseError transforms a [github.com/ory/keto/proto/ory/keto/opl/v1alpha1.ParseError] into a [ParseError].
func newParseError(parseError *opl.ParseError) ParseError {
	err := ParseError{
		Message: parseError.Message,
	}
	if parseError.Start != nil {
		err.StartPosition = &SourcePosition{
			Line:   parseError.Start.Line,
			Column: parseError.Start.Column,
		}
	}

	if parseError.End != nil {
		err.EndPosition = &SourcePosition{
			Line:   parseError.End.Line,
			Column: parseError.End.Column,
		}
	}

	return err
}

// Check checks if an OPL query's syntax is valid.
// It will return an error if the GRPC call failed OR if the syntax is incorrect.
// In such case, you can errors.As() against the [ParseError] type.
// nil indicates syntax is correct.
func (client *syntaxClient) Check(ctx context.Context, data []byte) error {
	req := opl.CheckRequest{Content: data}

	res, err := client.syntaxClient.Check(ctx, &req)
	if err != nil {
		return fmt.Errorf("grpc call: %w", err)
	}

	if len(res.ParseErrors) != 0 && res.ParseErrors[0] != nil {
		return newParseError(res.ParseErrors[0])
	}

	return nil
}
