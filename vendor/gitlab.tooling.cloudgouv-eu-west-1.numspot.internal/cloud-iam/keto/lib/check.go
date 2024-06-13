package lib

import (
	"context"
	"fmt"

	relations "github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2"
)

// Check whether a specific subject is related to an object.
func (client *readClient) Check(ctx context.Context, tuple RelationTuple, maxDepth *int32) (bool, error) {
	req := relations.CheckRequest{
		Tuple: relationTupleToGRPC(tuple),
	}

	if maxDepth != nil {
		req.MaxDepth = *maxDepth
	}

	res, err := client.checkServiceClient.Check(ctx, &req)
	if err != nil {
		return false, fmt.Errorf("grpc call: %w", err)
	}

	return res.GetAllowed(), nil
}
