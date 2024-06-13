package pkg

import "errors"

// NotifyAction enum.
type NotifyAction string

const (
	// NotifyActionCreate enum variant.
	NotifyActionCreate NotifyAction = "create"
	// NotifyActionDelete enum variant.
	NotifyActionDelete NotifyAction = "delete"
)

// ErrInvalidNotifyAction is returned when calling [NotifyAction.Validate].
var ErrInvalidNotifyAction = errors.New("invalid identity type")

// Validate validates an NotifyAction.
func (t NotifyAction) Validate() error {
	switch t {
	case NotifyActionCreate, NotifyActionDelete:
		return nil
	default:
		return ErrInvalidNotifyAction
	}
}
