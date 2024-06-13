package lib

// SubjectSet refers to all [Subjects](Subject) who have the same relation on an object.
type SubjectSet struct {
	// The namespace of the object and relation referenced in this subject set.
	Namespace Namespace
	// The object related by this subject set.
	Object string
	// The relation between the object and the subjects.
	Relation string
}
