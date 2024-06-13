package lib

import (
	"context"
	"fmt"
	"io"

	relations "github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2"
	"google.golang.org/grpc"
)

type readClient struct {
	conn                *grpc.ClientConn
	namespaceClient     relations.NamespacesServiceClient
	versionClient       relations.VersionServiceClient
	readServiceClient   relations.ReadServiceClient
	checkServiceClient  relations.CheckServiceClient
	expandServiceClient relations.ExpandServiceClient
}

// Reader used to communicate with the read services exported by Keto.
type Reader interface {
	io.Closer
	Expand(ctx context.Context, subject Subject, maxDepth *int32) (*SubjectTree, error)
	Check(ctx context.Context, tuple RelationTuple, maxDepth *int32) (bool, error)
	ListNamespaces(ctx context.Context) ([]Namespace, error)
	GetVersion(ctx context.Context) (string, error)
	ListRelationTuples(ctx context.Context, query *RelationQuery, pageSize *int32, cursor *string) ([]RelationTuple, string, error)
}

// Close underlying gRPC connection.
func (client *readClient) Close() error {
	if err := client.conn.Close(); err != nil {
		return fmt.Errorf("conn.Close: %w", err)
	}

	return nil
}

// NewReadClient initializes a [Reader] or fails if it couldn't [grpc.Dial] the read server.
func NewReadClient(host, port string, opts ...grpc.DialOption) (Reader, error) {
	var client readClient

	readClient, err := grpc.Dial(fmt.Sprintf("%s:%s", host, port), opts...)
	if err != nil {
		return nil, fmt.Errorf("grpc dial: %w", err)
	}

	client.conn = readClient

	client.expandServiceClient = relations.NewExpandServiceClient(readClient)
	client.namespaceClient = relations.NewNamespacesServiceClient(readClient)
	client.versionClient = relations.NewVersionServiceClient(readClient)
	client.readServiceClient = relations.NewReadServiceClient(readClient)
	client.checkServiceClient = relations.NewCheckServiceClient(readClient)

	return &client, nil
}
