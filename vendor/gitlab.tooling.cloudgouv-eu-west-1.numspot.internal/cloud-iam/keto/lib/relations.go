package lib

import (
	"context"
	"fmt"

	relations "github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2"
)

// ModifyRelationTuples creates and/or delete [RelationTuple]'s in keto.
func (client *writeClient) ModifyRelationTuples(ctx context.Context, tuples []RelationTupleDelta) error {
	req := &relations.TransactRelationTuplesRequest{}
	ketoTuples := make([]*relations.RelationTupleDelta, 0, len(tuples))
	for i := range tuples {
		ketoTuples = append(ketoTuples, &relations.RelationTupleDelta{
			Action:        client.getAction(tuples[i].Action),
			RelationTuple: relationTupleToGRPC(*tuples[i].RelationTuple),
		})
	}

	req.RelationTupleDeltas = ketoTuples

	_, err := client.writeServiceClient.TransactRelationTuples(ctx, req)
	if err != nil {
		return fmt.Errorf("grpc call: %w", err)
	}

	return nil
}

func (client *writeClient) getAction(action TupleAction) relations.RelationTupleDelta_Action {
	a := relations.RelationTupleDelta_ACTION_UNSPECIFIED
	switch action {
	case TupleActionInsert:
		a = relations.RelationTupleDelta_ACTION_INSERT
	case TupleActionDelete:
		a = relations.RelationTupleDelta_ACTION_DELETE
	}
	return a
}

// AddRelationTuples creates new [RelationTuple]'s in keto.
// They are all inserted in one transaction.
func (client *writeClient) AddRelationTuples(ctx context.Context, tuples []RelationTuple) error {
	req := &relations.TransactRelationTuplesRequest{}
	ketoTuples := make([]*relations.RelationTuple, 0, len(tuples))
	for i := range tuples {
		ketoTuples = append(ketoTuples, relationTupleToGRPC(tuples[i]))
	}

	req.RelationTupleDeltas = relations.RelationTupleToDeltas(ketoTuples, relations.RelationTupleDelta_ACTION_INSERT)

	_, err := client.writeServiceClient.TransactRelationTuples(ctx, req)
	if err != nil {
		return fmt.Errorf("grpc call: %w", err)
	}

	return nil
}

// DeleteRelationTuples removes [RelationTuple]'s from keto.
func (client *writeClient) DeleteRelationTuples(ctx context.Context, tuples []RelationTuple) error {
	req := &relations.TransactRelationTuplesRequest{}
	ketoTuples := make([]*relations.RelationTuple, 0, len(tuples))
	for i := range tuples {
		ketoTuples = append(ketoTuples, relationTupleToGRPC(tuples[i]))
	}

	req.RelationTupleDeltas = relations.RelationTupleToDeltas(ketoTuples, relations.RelationTupleDelta_ACTION_DELETE)

	_, err := client.writeServiceClient.TransactRelationTuples(ctx, req)
	if err != nil {
		return fmt.Errorf("grpc call: %w", err)
	}

	return nil
}

// DeleteRelationTuplesFromQuery removes [RelationTuple]'s from keto using a [RelationQuery].
func (client *writeClient) DeleteRelationTuplesFromQuery(ctx context.Context, query RelationQuery) error {
	req := &relations.DeleteRelationTuplesRequest{
		RelationQuery: relationQueryToGRPC(query),
	}

	_, err := client.writeServiceClient.DeleteRelationTuples(ctx, req)
	if err != nil {
		return fmt.Errorf("grpc call: %w", err)
	}

	return nil
}

// relationTupleToGRPC transforms a [RelationTuple] into a (keto.RelationTuple)[github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2.RelationTuple].
func relationTupleToGRPC(tuple RelationTuple) *relations.RelationTuple {
	return &relations.RelationTuple{
		Namespace: tuple.Namespace.String(),
		Object:    tuple.Object,
		Relation:  tuple.Relation,
		Subject:   subjectToGRPC(tuple.Subject),
	}
}

// relationTupleFromGRPC transforms a (keto.RelationTuple)[github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2.RelationTuple] into a [RelationTuple].
func relationTupleFromGRPC(set *relations.RelationTuple) *RelationTuple {
	tuple := &RelationTuple{
		Namespace: Namespace(set.GetNamespace()),
		Object:    set.GetObject(),
		Relation:  set.GetRelation(),
	}

	if subject := set.GetSubject(); subject != nil {
		tuple.Subject = *subjectFromGRPC(subject)
	}

	return tuple
}

// relationQueryToGRPC transforms a [RelationQuery] into a (keto.RelationQuery)[github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2.RelationQuery].
func relationQueryToGRPC(query RelationQuery) *relations.RelationQuery {
	var subject *relations.Subject
	if query.Subject != nil {
		subject = subjectToGRPC(*query.Subject)
	}

	var ns *string
	if query.Namespace != nil {
		nsTmp := query.Namespace.String()
		ns = &nsTmp
	}

	return &relations.RelationQuery{
		Namespace: ns,
		Object:    query.Object,
		Relation:  query.Relation,
		Subject:   subject,
	}
}
