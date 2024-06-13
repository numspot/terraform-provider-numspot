package lib

// NodeType of a [SubjectTree] node.
type NodeType int32

const (
	// NodeTypeUnspecified should not be received. It's a sentinel value.
	NodeTypeUnspecified NodeType = iota
	// NoteTypeUnion expands to a union of all children.
	NoteTypeUnion
	// NodeTypeExclusion is not implemented yet.
	NodeTypeExclusion
	// NodeTypeIntersection is not implemented yet.
	NodeTypeIntersection
	// NodeTypeLeaf contains no children.
	NodeTypeLeaf
)

// The String representation of the [NodeType] enum value.
func (nodeType NodeType) String() string {
	switch nodeType {
	case NodeTypeUnspecified:
		return "NODE_TYPE_UNSPECIFIED"
	case NoteTypeUnion:
		return "NODE_TYPE_UNION"
	case NodeTypeExclusion:
		return "NODE_TYPE_EXCLUSION"
	case NodeTypeIntersection:
		return "NODE_TYPE_INTERSECTION"
	case NodeTypeLeaf:
		return "NODE_TYPE_LEAF"
	default:
		return ""
	}
}
