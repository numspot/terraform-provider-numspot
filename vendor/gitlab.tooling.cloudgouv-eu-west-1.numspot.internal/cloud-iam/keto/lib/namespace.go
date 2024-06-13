package lib

import (
	"context"
	"fmt"

	relations "github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2"
)

// RootNodeID is a dummy constant used to identity the root node in everything.
const RootNodeID = "root"

// ListNamespaces returns a list of namespaces.
func (client *readClient) ListNamespaces(ctx context.Context) ([]Namespace, error) {
	req := relations.ListNamespacesRequest{}
	res, err := client.namespaceClient.ListNamespaces(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("grpc call: %w", err)
	}

	resNamespaces := res.GetNamespaces()

	namespaces := make([]Namespace, 0, len(resNamespaces))

	for i := range resNamespaces {
		if resNamespaces[i] != nil {
			namespaces = append(namespaces, Namespace(resNamespaces[i].GetName()))
		}
	}

	return namespaces, nil
}
