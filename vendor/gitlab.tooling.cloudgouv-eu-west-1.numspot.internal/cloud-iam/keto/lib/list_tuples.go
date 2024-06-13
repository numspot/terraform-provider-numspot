package lib

import (
	"context"
	"fmt"

	relations "github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2"
)

// ListRelationTuples returns a list of tuples corresponding to the query.
// It also returns the next page token.
func (client *readClient) ListRelationTuples(ctx context.Context, query *RelationQuery, pageSize *int32, cursor *string) ([]RelationTuple, string, error) {
	req := &relations.ListRelationTuplesRequest{}

	if query != nil {
		req.RelationQuery = relationQueryToGRPC(*query)
	}

	if pageSize != nil {
		req.PageSize = *pageSize
	}

	if cursor != nil {
		req.PageToken = *cursor
	}
	res, err := client.readServiceClient.ListRelationTuples(ctx, req)
	if err != nil {
		return nil, "", fmt.Errorf("grpc call: %w", err)
	}

	resRelationTuples := res.GetRelationTuples()

	tuples := make([]RelationTuple, 0, len(resRelationTuples))

	for i := range resRelationTuples {
		if resRelationTuples[i] != nil {
			tuples = append(tuples, *relationTupleFromGRPC(resRelationTuples[i]))
		}
	}

	return tuples, res.NextPageToken, nil
}
