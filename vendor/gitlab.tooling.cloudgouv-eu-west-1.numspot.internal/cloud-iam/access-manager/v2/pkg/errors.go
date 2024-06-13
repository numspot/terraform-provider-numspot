package pkg

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	keto "gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud-iam/keto/lib"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

var (
	// ErrUnauthorized signifies a user is not properly authenticated.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrPermissionNotFound returned to signify no corresponding permission was found.
	ErrPermissionNotFound = errors.New("permission not found")
)

// ObjectUnauthorizedError means that the object does not belong to the given tenant.
type ObjectUnauthorizedError struct {
	ObjectId   string
	ObjectType keto.Namespace
}

func (o ObjectUnauthorizedError) Error() string {
	return fmt.Sprintf("%s %s unauthorized", o.ObjectType, o.ObjectId)
}

// IdentityNotInTenantError means that the identity does not belong to the given tenant.
type IdentityNotInTenantError struct {
	IdentityId   uuid.UUID
	IdentityType IdentityType
	TenantId     uuid.UUID
}

// Error implements [builtin.error].
func (i IdentityNotInTenantError) Error() string {
	return fmt.Sprintf("identity %s of type %s does not belong to the given tenant %s", i.IdentityId, i.IdentityType, i.TenantId)
}

// ForbiddenError is returned when calling [verify.Verify].
type ForbiddenError struct {
	IdentityID   uuid.UUID
	PermissionID uuid.UUID
	ObjectID     *string
}

// Error string containing the incriminated IDs.
func (err ForbiddenError) Error() string {
	if err.ObjectID != nil {
		return fmt.Sprintf("user %s does not have permission %s on object %s", err.IdentityID, err.PermissionID, *err.ObjectID)
	}

	return fmt.Sprintf("user %s does not have permission %s", err.IdentityID, err.PermissionID)
}

// ForbiddenErrorFromGRPC will parse a [status.Status] into a [*ForbiddenError].
// This is a blind-ish conversion and doesn't check if the status should be considered a ForbiddenError.
func ForbiddenErrorFromGRPC(st *status.Status) (*ForbiddenError, error) {
	det := st.Details()
	if len(det) == 0 {
		return nil, errors.New("st.Details() is empty") //nolint:goerr113 // this does not need to be checkable against.
	}

	infos, ok := det[0].(*errdetails.ErrorInfo)
	if !ok {
		return nil, errors.New("st.Details() is not of type *errdetails.ErrorInfo") //nolint:goerr113 // this does not need to be checkable against.
	}

	var (
		forbiddenErr = new(ForbiddenError)
		err          error
	)

	forbiddenErr.IdentityID, err = uuid.Parse(infos.Metadata["identity_uuid"])
	if err != nil {
		return nil, fmt.Errorf("uuid.Parse(infos.Metadata[identity_uuid]): %w", err)
	}

	forbiddenErr.PermissionID, err = uuid.Parse(infos.Metadata["permission_uuid"])
	if err != nil {
		return nil, fmt.Errorf("uuid.Parse(infos.Metadata[permission_uuid]): %w", err)
	}

	if objectID := infos.Metadata["object"]; objectID != "" {
		forbiddenErr.ObjectID = &objectID
	}

	return forbiddenErr, nil
}

// BadRequestError is returned when calling [verify.Verify].
// It will contain the list of malformed fields.
type BadRequestError struct {
	Message string
	Details []struct {
		Field     string
		Violation string
	}
}

// Error string containing the details of malformed fields and what was wrong with them.
func (err BadRequestError) Error() string {
	builder := new(strings.Builder)
	builder.WriteString(err.Message)
	builder.WriteString(": ")
	for i := range err.Details {
		fmt.Fprintf(builder, "(%s: %s)", err.Details[i].Field, err.Details[i].Violation)
	}

	return builder.String()
}

// BadRequestErrorFromGRPC will parse a [status.Status] into a [*BadRequestError].
// This is a blind-ish conversion and doesn't check if the status should be considered a BadRequestError.
func BadRequestErrorFromGRPC(st *status.Status) *BadRequestError {
	badReqErr := new(BadRequestError)
	badReqErr.Message = st.Message()
	details := st.Details()
	for i := range details {
		detail, ok := details[i].(*errdetails.BadRequest_FieldViolation)
		if !ok {
			continue
		}

		errDetail := struct {
			Field     string
			Violation string
		}{
			Field:     detail.Field,
			Violation: detail.Description,
		}

		badReqErr.Details = append(badReqErr.Details, errDetail)
	}

	return badReqErr
}
