package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
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

	spaceId, err := uuid.Parse(plan.SpaceId.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid space_id", "space_id should be in UUID format")
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
	tf := CreateServiceAccountResponseFromHTTPToTF(*res.JSON201)
	tf.SpaceId = plan.SpaceId
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

	spaceId, err := uuid.Parse(state.SpaceId.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid space_id", "space_id should be in UUID format")
		return
	}

	serviceAccountID, err := uuid.Parse(state.ServiceAccountId.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid service_account_id", "service_account_id should be in UUID format")
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

	tf := ServiceAccountEditedResponseFromHTTPToTF(*res.JSON200)
	state.ServiceAccountId = tf.ServiceAccountId
	state.Name = tf.Name
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (r *ServiceAccountResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state resource_service_account.ServiceAccountModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	spaceId, err := uuid.Parse(plan.SpaceId.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid space_id", "space_id should be in UUID format")
		return
	}

	serviceAccountID, err := uuid.Parse(state.ServiceAccountId.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid service_account_id", "service_account_id should be in UUID format")
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

	tf := ServiceAccountEditedResponseFromHTTPToTF(*res.JSON200)
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

	spaceId, err := uuid.Parse(state.SpaceId.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid space_id", "space_id should be in UUID format")
		return
	}

	serviceAccountID, err := uuid.Parse(state.ServiceAccountId.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid service_account_id", "service_account_id should be in UUID format")
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
