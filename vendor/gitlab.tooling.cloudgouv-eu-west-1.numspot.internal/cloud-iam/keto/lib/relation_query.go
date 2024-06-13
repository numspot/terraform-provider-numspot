package lib

// The RelationQuery for listing relationships. Clients can specify any optional field to partially filter for specific relationships.
//
// Example use cases (namespace is always required):
//
//   - object only: display a list of all permissions referring to a specific object
//   - relation only: get all groups that have members; get all directories that have content
//   - object & relation: display all subjects that have a specific permission relation
//   - subject & relation: display all groups a subject belongs to; display all objects a subject has access to
//   - object & relation & subject: verify whether the relation tuple already exists
type RelationQuery struct {
	// The namespace this relation tuple lives in.
	Namespace *Namespace
	// The object related by this tuple. It is an object in the namespace of the tuple.
	Object *string
	// 	The relation between an Object and a Subject.
	Relation *string
	// The subject related by this tuple. A Subject either represents a concrete subject id or a SubjectSet that expands to more Subjects.
	Subject *Subject
}
