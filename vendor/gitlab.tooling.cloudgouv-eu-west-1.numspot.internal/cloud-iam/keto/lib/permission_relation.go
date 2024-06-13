package lib

import "errors"

// PermissionRelation to an object.
type PermissionRelation string

// ErrInvalidPermissionRelation is returned when calling [PermissionRelation.Validate] or [PermissionRelation.PermissionRelation].
var ErrInvalidPermissionRelation = errors.New("invalid permission relation")

// String representation.
// Empty string in case relation is invalid.
func (relation PermissionRelation) String() string {
	if relation.Validate() == nil {
		return string(relation)
	}

	return ""
}

// PermissionAction linked to a PermissionRelation.
func (relation PermissionRelation) PermissionAction() (PermissionAction, error) {
	for a, r := range actionRelation {
		if r == relation {
			return a, nil
		}
	}

	return "", ErrInvalidPermissionRelation
}

// Validate enum value.
func (relation PermissionRelation) Validate() error {
	for _, r := range actionRelation {
		if r == relation {
			return nil
		}
	}

	return ErrInvalidPermissionRelation
}
