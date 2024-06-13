package lib

import (
	"context"
	"fmt"
	"io"

	relations "github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2"
	"google.golang.org/grpc"
)

type writeClient struct {
	conn               *grpc.ClientConn
	writeServiceClient relations.WriteServiceClient
}

// Writer used to communicate with the write services exported by Keto.
type Writer interface {
	io.Closer
	AddRelationTuples(ctx context.Context, tuples []RelationTuple) error
	DeleteRelationTuples(ctx context.Context, tuples []RelationTuple) error
	ModifyRelationTuples(ctx context.Context, tuples []RelationTupleDelta) error
	DeleteRelationTuplesFromQuery(ctx context.Context, query RelationQuery) error
}

// Close underlying gRPC connection.
func (client *writeClient) Close() error {
	if err := client.conn.Close(); err != nil {
		return fmt.Errorf("conn.Close: %w", err)
	}

	return nil
}

// NewWriteClient initializes a [Writer] or fails if it couldn't [grpc.Dial] the write server.
func NewWriteClient(host, port string, opts ...grpc.DialOption) (Writer, error) {
	var client writeClient

	writeClient, err := grpc.Dial(fmt.Sprintf("%s:%s", host, port), opts...)
	if err != nil {
		return nil, fmt.Errorf("grpc dial: %w", err)
	}

	client.conn = writeClient

	client.writeServiceClient = relations.NewWriteServiceClient(writeClient)

	return &client, nil
}
