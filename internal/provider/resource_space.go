package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_space"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &SpaceResource{}
	_ resource.ResourceWithConfigure   = &SpaceResource{}
	_ resource.ResourceWithImportState = &SpaceResource{}
)

type SpaceResource struct {
	provider Provider
}

func NewSpaceResource() resource.Resource {
	return &SpaceResource{}
}

func (r *SpaceResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *SpaceResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *SpaceResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_space"
}

func (r *SpaceResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_space.SpaceResourceSchema(ctx)
}

func (r *SpaceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_space.SpaceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	organisationId, err := uuid.Parse(plan.OrganisationId.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid organisation_id", "organisation_id should be in UUID format")
		return
	}
	res := utils.ExecuteRequest(func() (*numspot.CreateSpaceResponse, error) {
		return r.provider.NumSpotClient.CreateSpaceWithResponse(
			ctx,
			organisationId,
			SpaceFromTfToCreateRequest(&plan),
		)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	readRes, err := RetryReadSpaceUntilReady(ctx, r.provider.NumSpotClient, res.JSON200.Id)
	if err != nil {
		response.Diagnostics.AddError("failed to read space", err.Error())
		return
	}

	space, ok := readRes.(*numspot.Space)
	if !ok {
		response.Diagnostics.AddError("failed to read space", "")
		return
	}

	tf := SpaceFromHttpToTf(space)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SpaceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_space.SpaceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	spaceId, err := uuid.Parse(data.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid space_id", "space_id should be in UUID format")
		return
	}
	res := utils.ExecuteRequest(func() (*numspot.GetSpaceByIdResponse, error) {
		return r.provider.NumSpotClient.GetSpaceByIdWithResponse(ctx, spaceId)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := SpaceFromHttpToTf(res.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SpaceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	panic("not implemented")
}

func (r *SpaceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_space.SpaceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	spaceId, err := uuid.Parse(data.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid space_id", "space_id should be in UUID format")
		return
	}
	res := utils.ExecuteRequest(func() (*numspot.DeleteSpaceResponse, error) {
		return r.provider.NumSpotClient.DeleteSpaceWithResponse(ctx, spaceId)
	}, http.StatusNoContent, &response.Diagnostics)
	if res == nil {
		return
	}

	response.State.RemoveResource(ctx)
}
