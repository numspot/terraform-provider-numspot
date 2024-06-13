package lib

import "errors"

// PermissionAction of a permission.
type PermissionAction string

// ErrInvalidPermissionAction is returned when calling [PermissionAction.Validate] or [PermissionAction.PermissionRelation].
var ErrInvalidPermissionAction = errors.New("invalid permission action")

// String representation.
// Empty string in case action is invalid.
func (action PermissionAction) String() string {
	if action.Validate() == nil {
		return string(action)
	}

	return ""
}

// PermissionRelation linked to a PermissionAction.
func (action PermissionAction) PermissionRelation() (PermissionRelation, error) {
	relation, ok := actionRelation[action]
	if !ok {
		return "", ErrInvalidPermissionAction
	}

	return relation, nil
}

// Validate enum value.
func (action PermissionAction) Validate() error {
	if _, ok := actionRelation[action]; !ok {
		return ErrInvalidPermissionAction
	}

	return nil
}
