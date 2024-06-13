package lib

import (
	"context"
	"fmt"

	relations "github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2"
)

// subjectTreeFromGRPC transforms a (keto.SubjectTree)[github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2.SubjectTree] into a [model.SubjectTree].
// ⚠️ BEWARE OF THE WILD RECURSIVE FUNCTION. ⚠️
// USE AT YOUR OWN RISK AND PERILS.
func subjectTreeFromGRPC(node *relations.SubjectTree) *SubjectTree {
	tree := &SubjectTree{
		NodeType: NodeType(node.GetNodeType()),
	}

	if tuple := node.GetTuple(); tuple != nil {
		tree.Tuple = relationTupleFromGRPC(tuple)
	}

	resChildren := node.GetChildren()
	children := make([]SubjectTree, 0, len(resChildren))
	for i := range resChildren {
		if resChildren[i] != nil {
			children = append(children, *subjectTreeFromGRPC(resChildren[i]))
		}
	}

	tree.Children = children

	return tree
}

// Expand the given subject set.
func (client *readClient) Expand(ctx context.Context, subject Subject, maxDepth *int32) (*SubjectTree, error) {
	req := relations.ExpandRequest{}

	req.Subject = subjectToGRPC(subject)
	if maxDepth != nil {
		req.MaxDepth = *maxDepth
	}

	res, err := client.expandServiceClient.Expand(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("grpc call: %w", err)
	}

	tree := subjectTreeFromGRPC(res.GetTree())

	return tree, nil
}
