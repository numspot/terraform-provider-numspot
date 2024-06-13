package pkg

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/orsinium-labs/enum"
)

// IdentityPathType model.
type IdentityPathType enum.Member[string]

// ToIdentity returns the [IdentityType].
func (i *IdentityPathType) ToIdentity() IdentityType {
	switch *i {
	case IdentityPathTypeService:
		return IdentityTypeService
	case IdentityPathTypeUser:
		return IdentityTypeUser
	default:
		return IdentityTypeInvalid
	}
}

// UnmarshalText implements [encoding.TextUnmarshaler] to allow the generated code to use that type.
func (i *IdentityPathType) UnmarshalText(text []byte) error {
	given := string(text)
	parse := IdentityPathTypes.Parse(given)
	if parse == nil || !IdentityPathTypes.Contains(*parse) {
		return &InvalidIdentityPathTypeError{given: given}
	}
	i.Value = parse.Value
	return nil
}

var (
	// IdentityPathTypeUser represent a user.
	IdentityPathTypeUser = IdentityPathType{Value: "users"}
	// IdentityPathTypeService represent a service account.
	IdentityPathTypeService = IdentityPathType{Value: "serviceAccounts"}
	// IdentityPathTypes represent all the valid values of IdentityPathType.
	IdentityPathTypes = enum.New(IdentityPathTypeUser, IdentityPathTypeService)
)

// InvalidIdentityPathTypeError is returned when the Cardinality is not valid.
type InvalidIdentityPathTypeError struct {
	given string
}

func (i *InvalidIdentityPathTypeError) Error() string {
	return fmt.Sprintf("%s is not a valid identity path, valid values are %s", i.given, IdentityPathTypes.Members())
}

// IdentityPathNotFoundError is returned when an identity is not found.
type IdentityPathNotFoundError struct {
	Ty IdentityPathType
	Id uuid.UUID
}

func (i IdentityPathNotFoundError) Error() string {
	return fmt.Sprintf("identity %s of type %s not found", i.Id.String(), i.Ty)
}
