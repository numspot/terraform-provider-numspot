package lib

import (
	relations "github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2"
)

// Subject is either a concrete subject id or a SubjectSet expanding to more [Subjects](Subject).
type Subject struct {
	ID  string
	Set *SubjectSet
}

// subjectFromGRPC transforms a (keto.Subject)[github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2.Subject] into a [Subject].
func subjectFromGRPC(subject *relations.Subject) *Subject {
	var set *SubjectSet
	if s := subject.GetSet(); s != nil {
		set = subjectSetFromGRPC(s)
	}
	return &Subject{
		ID:  subject.GetId(),
		Set: set,
	}
}

// subjectSetToGRPC transforms a [Subject] into a (keto.Subject)[github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2.Subject].
func subjectToGRPC(subject Subject) *relations.Subject {
	if subject.Set != nil {
		return &relations.Subject{
			Ref: &relations.Subject_Set{Set: subjectSetToGRPC(*subject.Set)},
		}
	}

	return &relations.Subject{
		Ref: &relations.Subject_Id{Id: subject.ID},
	}
}

// subjectSetFromGRPC transforms a (SubjectSet)[github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2.SubjectSet] into a [SubjectSet].
func subjectSetFromGRPC(set *relations.SubjectSet) *SubjectSet {
	return &SubjectSet{
		Namespace: Namespace(set.GetNamespace()),
		Object:    set.GetObject(),
		Relation:  set.GetRelation(),
	}
}

// subjectSetToGRPC transforms a [SubjectSet] into a (keto.SubjectSet)[github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2.SubjectSet].
func subjectSetToGRPC(set SubjectSet) *relations.SubjectSet {
	return &relations.SubjectSet{
		Namespace: set.Namespace.String(),
		Object:    set.Object,
		Relation:  set.Relation,
	}
}
