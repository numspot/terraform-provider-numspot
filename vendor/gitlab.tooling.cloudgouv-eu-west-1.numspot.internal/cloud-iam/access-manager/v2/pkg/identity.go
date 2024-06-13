package pkg

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/orsinium-labs/enum"
)

// IdentityType model.
type IdentityType enum.Member[string]

// UnmarshalJSON implements Unmarshaler.
func (c *IdentityType) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &c.Value)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}
	if !IdentityTypesDeprecated.Contains(*c) {
		return &InvalidIdentityTypeError{given: string(b)}
	}
	*c = c.ToNotDeprecated()
	return nil
}

// ToNotDeprecated returns the not deprecated equivalent.
// Deprecated : to be removed when the migration is complete.
func (c *IdentityType) ToNotDeprecated() IdentityType {
	switch *c {
	case IdentityTypeUser, IdentityTypeUserDeprecated:
		return IdentityTypeUser
	case IdentityTypeService, IdentityTypeServiceDeprecated:
		return IdentityTypeService
	default:
		return IdentityTypeInvalid
	}
}

var (
	// IdentityTypeUser represent a user.
	IdentityTypeUser = IdentityType{Value: "user"}
	// IdentityTypeUserDeprecated represent a user.
	// Deprecated : use [IdentityTypeUser]
	IdentityTypeUserDeprecated = IdentityType{Value: "users"}
	// IdentityTypeService represent a service account.
	IdentityTypeService = IdentityType{Value: "serviceAccount"}
	// IdentityTypeServiceDeprecated represent a service account.
	// Deprecated : use [IdentityTypeService]
	IdentityTypeServiceDeprecated = IdentityType{Value: "serviceAccounts"}
	// IdentityTypeInvalid represent in invalid value for an IdentityType.
	IdentityTypeInvalid = IdentityType{Value: "invalid"}
	// IdentityTypesDeprecated represent all the valid values of IdentityType.
	IdentityTypesDeprecated = enum.New(IdentityTypeUser, IdentityTypeUserDeprecated, IdentityTypeService, IdentityTypeServiceDeprecated)
	// IdentityTypes represent all the valid values of IdentityType.
	IdentityTypes = enum.New(IdentityTypeUser, IdentityTypeService)
)

// InvalidIdentityTypeError is returned when the Cardinality is not valid.
type InvalidIdentityTypeError struct {
	given string
}

func (i *InvalidIdentityTypeError) Error() string {
	return fmt.Sprintf("%s is not a valid entity type, valid values are %s", i.given, IdentityTypes.Members())
}

// ErrInvalidIdentityType is returned when calling [IdentityType.Validate].
var ErrInvalidIdentityType = errors.New("invalid identity type")

// IdentityNotFoundError is returned when an identity is not found.
type IdentityNotFoundError struct {
	Ty IdentityType
	Id uuid.UUID
}

func (i IdentityNotFoundError) Error() string {
	return fmt.Sprintf("identity %s of type %s not found", i.Id.String(), i.Ty)
}
