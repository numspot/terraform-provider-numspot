package lib

import (
	"context"
	"fmt"
	"io"

	opl "github.com/ory/keto/proto/ory/keto/opl/v1alpha1"
	"google.golang.org/grpc"
)

type syntaxClient struct {
	conn         *grpc.ClientConn
	syntaxClient opl.SyntaxServiceClient
}

// SyntaxChecker used to communicate with the syntax services exported by Keto.
type SyntaxChecker interface {
	io.Closer
	Check(ctx context.Context, data []byte) error
}

// Close underlying gRPC connection.
func (client *syntaxClient) Close() error {
	if err := client.conn.Close(); err != nil {
		return fmt.Errorf("conn.Close: %w", err)
	}

	return nil
}

// NewSyntaxClient initializes a [SyntaxChecker] or fails if it couldn't [grpc.Dial] the syntax server.
func NewSyntaxClient(host, port string, opts ...grpc.DialOption) (SyntaxChecker, error) {
	var client syntaxClient

	syntaxClient, err := grpc.Dial(fmt.Sprintf("%s:%s", host, port), opts...)
	if err != nil {
		return nil, fmt.Errorf("grpc dial: %w", err)
	}

	client.conn = syntaxClient

	client.syntaxClient = opl.NewSyntaxServiceClient(syntaxClient)

	return &client, nil
}
