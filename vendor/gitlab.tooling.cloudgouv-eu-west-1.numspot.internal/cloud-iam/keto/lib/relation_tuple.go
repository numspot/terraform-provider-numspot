package lib

import (
	"fmt"
	"strings"
)

// RelationTuple defines a relation between an Object and a Subject.
type RelationTuple struct {
	// The namespace this relation tuple lives in.
	Namespace Namespace
	// The object related by this tuple. It is an object in the namespace of the tuple.
	Object string
	// The relation between an Object and a Subject.
	Relation string
	// The subject related by this tuple. A Subject either represents a concrete subject id or a SubjectSet that expands to more Subjects.
	Subject Subject
}

// String representation.
func (tuple RelationTuple) String() string {
	b := new(strings.Builder)
	fmt.Fprintf(b, "%s:%s", tuple.Namespace, tuple.Object)
	if tuple.Relation != "" {
		fmt.Fprintf(b, "#%s", tuple.Relation)
	}
	if tuple.Subject.ID != "" {
		fmt.Fprintf(b, "@(%s)", tuple.Subject.ID)
	} else if tuple.Subject.Set != nil {
		fmt.Fprintf(b, "@(%s:%s", tuple.Subject.Set.Namespace, tuple.Subject.Set.Object)
		if tuple.Subject.Set.Relation != "" {
			fmt.Fprintf(b, "#%s", tuple.Relation)
		}
		b.WriteRune(')')
	}

	return b.String()
}

// TupleAction represents an action type on a tuple.
type TupleAction string

// TupleAction* represents an action on a tuple.
const (
	TupleActionInsert TupleAction = "insert"
	TupleActionDelete TupleAction = "delete"
)

// RelationTupleDelta defines an action on a tuple.
type RelationTupleDelta struct {
	Action        TupleAction
	RelationTuple *RelationTuple
}
