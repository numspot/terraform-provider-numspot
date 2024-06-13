package lib

// SubjectTree node.
type SubjectTree struct {
	NodeType NodeType
	Tuple    *RelationTuple
	Children []SubjectTree
}
