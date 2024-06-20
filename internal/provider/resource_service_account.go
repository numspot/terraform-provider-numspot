package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iam"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_service_account"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &ServiceAccountResource{}
	_ resource.ResourceWithConfigure   = &ServiceAccountResource{}
	_ resource.ResourceWithImportState = &ServiceAccountResource{}
)

type (
	EntityType string
	Action     string
)

const (
	EntityTypePermission     EntityType = "permission"
	EntityTypeRole           EntityType = "role"
	EntityTypeSpace          EntityType = "space"
	EntityTypeServiceAccount EntityType = "service account"
)

const (
	AddAction    Action = "add"
	DeleteAction Action = "delete"
)

type ServiceAccountResource struct {
	provider Provider
}

func NewServiceAccountResource() resource.Resource {
	return &ServiceAccountResource{}
}

func (r *ServiceAccountResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(Provider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.provider = provider
}

func (r *ServiceAccountResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	idParts := strings.Split(request.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: space_id,service_account_id. Got: %q", request.ID),
		)
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("space_id"), idParts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("service_account_id"), idParts[1])...)
}

func (r *ServiceAccountResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_service_account"
}

func (r *ServiceAccountResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_service_account.ServiceAccountResourceSchema(ctx)
}

func (r *ServiceAccountResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_service_account.ServiceAccountModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	spaceId, diags := utils.ParseUUID(plan.SpaceId.ValueString())
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	res := utils.ExecuteRequest(func() (*iam.CreateServiceAccountSpaceResponse, error) {
		return r.provider.IAMIdentityManagerClient.CreateServiceAccountSpaceWithResponse(
			ctx,
			spaceId,
			ServiceAccountFromTFToCreateRequest(plan),
		)
	}, http.StatusCreated, &response.Diagnostics)
	if res == nil {
		return
	}

	if res.JSON201 == nil {
		response.Diagnostics.AddError("failed to create service account", "empty response")
		return
	}

	tf := CreateServiceAccountResponseFromHTTPToTF(ctx, *res.JSON201)
	tf.SpaceId = plan.SpaceId

	// This var is used to know if we need to fetch roles & permissions after creation
	verifyRolesAndPermissions := false

	// Attach permissions
	if len(plan.GlobalPermissions.Elements()) > 0 {
		globalPermissions := utils.FromTfStringSetToStringList(ctx, plan.GlobalPermissions)
		diags := r.modifyServiceAccountIAMPolicy(ctx, AddAction, EntityTypePermission, spaceId, res.JSON201.Id, globalPermissions)
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}

		verifyRolesAndPermissions = true
	}

	// Attach Roles
	if len(plan.Roles.Elements()) > 0 {
		roles := utils.FromTfStringSetToStringList(ctx, plan.Roles)
		diags := r.modifyServiceAccountIAMPolicy(ctx, AddAction, EntityTypeRole, spaceId, res.JSON201.Id, roles)
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}

		verifyRolesAndPermissions = true
	}

	if verifyRolesAndPermissions {
		roles, permissions, diags := r.getRolesAndGlobalPermissions(ctx, res.JSON201.Id)
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}

		if roles != nil && len(*roles) > 0 {
			rolesTf, diags := utils.FromUUIDListToTfStringSet(ctx, *roles)
			if diags.HasError() {
				response.Diagnostics.Append(diags...)
				return
			}
			tf.Roles = rolesTf
		}

		if permissions != nil && len(*permissions) > 0 {
			permissionsTf, diags := utils.FromUUIDListToTfStringSet(ctx, *permissions)
			if diags.HasError() {
				response.Diagnostics.Append(diags...)
				return
			}
			tf.GlobalPermissions = permissionsTf
		}
	}

	// Read operation requires space_id and service_account_id to be able to fetch data from the API
	// As the import state operation will take into consideration only one ID attribute,
	// We have to combine the two attributes in a single ID attribute
	// Convention for combined ID is: space_id, service_account_id
	tf.Id = types.StringValue(fmt.Sprintf("%s,%s", tf.SpaceId.ValueString(), tf.Id.ValueString()))

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *ServiceAccountResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_service_account.ServiceAccountModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	spaceId, diags := utils.ParseUUID(state.SpaceId.ValueString())
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	serviceAccountID, diags := utils.ParseUUID(state.ServiceAccountId.ValueString())
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	res := utils.ExecuteRequest(func() (*iam.GetServiceAccountSpaceResponse, error) {
		return r.provider.IAMIdentityManagerClient.GetServiceAccountSpaceWithResponse(ctx, spaceId, serviceAccountID)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	if res.JSON200 == nil {
		response.Diagnostics.AddError("failed to read service account", "empty response")
		return
	}

	roles, permissions, diags := r.getRolesAndGlobalPermissions(ctx, state.ServiceAccountId.ValueString())
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	if roles != nil && len(*roles) > 0 {
		rolesTf, diags := utils.FromUUIDListToTfStringSet(ctx, *roles)
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}
		state.Roles = rolesTf
	}

	if permissions != nil && len(*permissions) > 0 {
		permissionsTf, diags := utils.FromUUIDListToTfStringSet(ctx, *permissions)
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}
		state.GlobalPermissions = permissionsTf
	}

	tf := ServiceAccountEditedResponseFromHTTPToTF(ctx, *res.JSON200)
	state.ServiceAccountId = tf.ServiceAccountId
	state.Name = tf.Name
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (r *ServiceAccountResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state resource_service_account.ServiceAccountModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	spaceId, diags := utils.ParseUUID(plan.SpaceId.ValueString())
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	serviceAccountID, diags := utils.ParseUUID(state.ServiceAccountId.ValueString())
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	if !plan.Name.Equal(state.Name) {
		tf, err := r.updateServiceAccount(ctx, spaceId, serviceAccountID, plan.Name.ValueString(), response)
		if err != nil {
			response.Diagnostics.AddError(err.Error(), "")
		}

		state.ServiceAccountId = tf.ServiceAccountId
		state.Name = plan.Name
		response.Diagnostics.Append(response.State.Set(ctx, state)...)
	}

	if !plan.SpaceId.Equal(state.SpaceId) {
		err := r.assignServiceAccountToSpace(ctx, spaceId, serviceAccountID, response)
		if err != nil {
			response.Diagnostics.AddError(err.Error(), "")
			return
		}
		state.SpaceId = plan.SpaceId
		response.Diagnostics.Append(response.State.Set(ctx, state)...)
	}

	// Roles & Permissions
	verifyRolesAndPermissions := false
	if !plan.GlobalPermissions.Equal(state.GlobalPermissions) {
		statePermissions := utils.FromTfStringSetToStringList(ctx, state.GlobalPermissions)
		planPermissions := utils.FromTfStringSetToStringList(ctx, plan.GlobalPermissions)

		diags := r.updateRolesOrPermission(ctx, spaceId, state.ServiceAccountId.ValueString(), EntityTypePermission, statePermissions, planPermissions)
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}

		verifyRolesAndPermissions = true
	}

	if !plan.Roles.Equal(state.Roles) {
		stateRoles := utils.FromTfStringSetToStringList(ctx, state.Roles)
		planRoles := utils.FromTfStringSetToStringList(ctx, plan.Roles)

		diags := r.updateRolesOrPermission(ctx, spaceId, state.ServiceAccountId.ValueString(), EntityTypeRole, stateRoles, planRoles)
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}

		verifyRolesAndPermissions = true
	}

	if verifyRolesAndPermissions {
		roles, permissions, diags := r.getRolesAndGlobalPermissions(ctx, state.ServiceAccountId.ValueString())
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}

		if roles != nil && len(*roles) > 0 {
			rolesTf, diags := utils.FromUUIDListToTfStringSet(ctx, *roles)
			if diags.HasError() {
				response.Diagnostics.Append(diags...)
				return
			}
			state.Roles = rolesTf
		} else {
			state.Roles = plan.Roles
		}

		if permissions != nil && len(*permissions) > 0 {
			permissionsTf, diags := utils.FromUUIDListToTfStringSet(ctx, *permissions)
			if diags.HasError() {
				response.Diagnostics.Append(diags...)
				return
			}
			state.GlobalPermissions = permissionsTf
		} else {
			state.GlobalPermissions = plan.GlobalPermissions
		}

		response.Diagnostics.Append(response.State.Set(ctx, state)...)
	}
}

func (r *ServiceAccountResource) updateServiceAccount(
	ctx context.Context,
	spaceID, servicAccountID uuid.UUID,
	serviceAccountName string,
	response *resource.UpdateResponse,
) (*resource_service_account.ServiceAccountModel, error) {
	payload := iam.ServiceAccount{Name: serviceAccountName}

	res := utils.ExecuteRequest(func() (*iam.UpdateServiceAccountSpaceResponse, error) {
		return r.provider.IAMIdentityManagerClient.UpdateServiceAccountSpaceWithResponse(ctx, spaceID, servicAccountID, payload)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return nil, fmt.Errorf("failed to update service account: %v", "empty response")
	}

	if res.JSON200 == nil {
		return nil, fmt.Errorf("failed to update service account: %v", "empty response")
	}

	tf := ServiceAccountEditedResponseFromHTTPToTF(ctx, *res.JSON200)
	return &tf, nil
}

func (r *ServiceAccountResource) assignServiceAccountToSpace(
	ctx context.Context,
	spaceID, servicAccountID uuid.UUID,
	response *resource.UpdateResponse,
) error {
	res := utils.ExecuteRequest(func() (*iam.AssignServiceAccountToSpaceResponse, error) {
		return r.provider.IAMIdentityManagerClient.AssignServiceAccountToSpaceWithResponse(ctx, spaceID, servicAccountID)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return fmt.Errorf("failed to assign service account to space: %v", "empty response")
	}

	return nil
}

func (r *ServiceAccountResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_service_account.ServiceAccountModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	spaceId, diags := utils.ParseUUID(state.SpaceId.ValueString())
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	serviceAccountID, diags := utils.ParseUUID(state.ServiceAccountId.ValueString())
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	res := utils.ExecuteRequest(func() (*iam.DeleteServiceAccountSpaceResponse, error) {
		return r.provider.IAMIdentityManagerClient.DeleteServiceAccountSpaceWithResponse(ctx, spaceId, serviceAccountID)
	}, http.StatusNoContent, &response.Diagnostics)
	if res == nil {
		return
	}

	response.State.RemoveResource(ctx)
}

func parseStringListAsUUIDs(strList []string) ([]uuid.UUID, diag.Diagnostics) {
	uuidBulk := make([]uuid.UUID, 0, len(strList))
	for _, str := range strList {
		currentUuid, diags := utils.ParseUUID(str)
		if diags.HasError() {
			return nil, diags
		}

		uuidBulk = append(uuidBulk, currentUuid)
	}

	return uuidBulk, nil
}

// Global Permissions & Roles
func (r *ServiceAccountResource) modifyServiceAccountIAMPolicy(
	ctx context.Context,
	action Action,
	entityType EntityType,
	spaceId uuid.UUID,
	serviceAccountID string,
	uuids []string,
) diag.Diagnostics {
	var diags diag.Diagnostics

	// Parse Service Account ID
	serviceAccountUUID, diags := utils.ParseUUID(serviceAccountID)
	if diags.HasError() {
		return diags
	}

	// Parse UUIDs
	uuidList, diags := parseStringListAsUUIDs(uuids)
	if diags.HasError() {
		return diags
	}

	// Create Body
	var policies *iam.IAMPolicy
	if entityType == EntityTypeRole {
		policies = &iam.IAMPolicy{Roles: &uuidList}
	} else if entityType == EntityTypePermission {
		policies = &iam.IAMPolicy{Permissions: &uuidList}
	}

	var body iam.SetIAMPolicySpaceJSONRequestBody
	if action == AddAction {
		body = iam.SetIAMPolicySpaceJSONRequestBody{
			Add: policies,
		}
	} else if action == DeleteAction {
		body = iam.SetIAMPolicySpaceJSONRequestBody{
			Delete: policies,
		}
	}

	// Execute
	utils.ExecuteRequest(func() (*iam.SetIAMPolicySpaceResponse, error) {
		return r.provider.IAMAccessManagerClient.SetIAMPolicySpaceWithResponse(
			ctx,
			spaceId,
			iam.ServiceAccounts,
			serviceAccountUUID,
			body,
		)
	}, http.StatusNoContent, &diags)

	return diags
}

func (r *ServiceAccountResource) updateRolesOrPermission(
	ctx context.Context,
	spaceId uuid.UUID,
	serviceAccountID string,
	entityType EntityType,
	stateValues, planValues []string,
) diag.Diagnostics {
	var diags diag.Diagnostics

	toAdd, toRemove := utils.DiffComparable(stateValues, planValues)

	if len(toRemove) > 0 {
		diags = r.modifyServiceAccountIAMPolicy(ctx, DeleteAction, entityType, spaceId, serviceAccountID, toRemove)
		if diags.HasError() {
			return diags
		}
	}

	if len(toAdd) > 0 {
		diags = r.modifyServiceAccountIAMPolicy(ctx, AddAction, entityType, spaceId, serviceAccountID, toAdd)
		if diags.HasError() {
			return diags
		}
	}

	return diags
}

func (r *ServiceAccountResource) getRolesAndGlobalPermissions(
	ctx context.Context,
	serviceAccountID string,
) (*[]uuid.UUID, *[]uuid.UUID, diag.Diagnostics) {
	var diags diag.Diagnostics

	serviceAccountUUID, diags := utils.ParseUUID(serviceAccountID)
	if diags.HasError() {
		return nil, nil, diags
	}

	res := utils.ExecuteRequest(func() (*iam.GetIAMPolicySpaceResponse, error) {
		return r.provider.IAMAccessManagerClient.GetIAMPolicySpaceWithResponse(
			ctx, r.provider.SpaceID, iam.ServiceAccounts, serviceAccountUUID)
	}, http.StatusOK, &diags)
	if res == nil {
		return nil, nil, diags
	}

	if res.JSON200 == nil {
		diags.AddError("Failed to get IAM policy space response", res.Status())
		return nil, nil, diags
	}

	roles := res.JSON200.Roles
	permissions := res.JSON200.Permissions

	return roles, permissions, nil
}
