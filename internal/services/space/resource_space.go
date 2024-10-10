package space

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &SpaceResource{}
	_ resource.ResourceWithConfigure   = &SpaceResource{}
	_ resource.ResourceWithImportState = &SpaceResource{}
)

type SpaceResource struct {
	provider *client.NumSpotSDK
}

func NewSpaceResource() resource.Resource {
	return &SpaceResource{}
}

func (r *SpaceResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(*client.NumSpotSDK)
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
	idParts := strings.Split(request.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: space_id,service_account_id. Got: %q", request.ID),
		)
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("organisation_id"), idParts[0])...)
	// response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("space_id"), idParts[1])...)
}

func (r *SpaceResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_space"
}

func (r *SpaceResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = SpaceResourceSchema(ctx)
}

func (r *SpaceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan SpaceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	// Retries create until request response is OK
	organisationId, err := uuid.Parse(plan.OrganisationId.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid organisation_id", "organisation_id should be in UUID format")
		return
	}
	res := utils.ExecuteRequest(func() (*numspot.CreateSpaceResponse, error) {
		return numspotClient.CreateSpaceWithResponse(
			ctx,
			organisationId,
			SpaceFromTfToCreateRequest(&plan),
		)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	readRes, err := RetryReadSpaceUntilReady(ctx, numspotClient, organisationId, res.JSON200.Id)
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

	// Read operation requires organisation_id_id and space_id to be able to fetch data from the API
	// As the import state operation will take into consideration only one ID attribute,
	// We have to combine the two attributes in a single ID attribute
	// Convention for combined ID is: organisation_id, space_id
	tf.SpaceId = tf.Id
	tf.Id = types.StringValue(fmt.Sprintf("%s,%s", tf.OrganisationId.ValueString(), tf.Id.ValueString()))

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SpaceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data SpaceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	spaceId, err := uuid.Parse(data.SpaceId.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid space_id", "space_id should be in UUID format")
		return
	}

	organisationId, err := uuid.Parse(data.OrganisationId.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid organisation_id", "organisation_id should be in UUID format")
		return
	}

	res := utils.ExecuteRequest(func() (*numspot.GetSpaceByIdResponse, error) {
		return numspotClient.GetSpaceByIdWithResponse(ctx, organisationId, spaceId)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := SpaceFromHttpToTf(res.JSON200)
	tf.Id = types.StringValue(fmt.Sprintf("%s,%s", tf.OrganisationId.ValueString(), tf.SpaceId.ValueString()))
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SpaceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	panic("not implemented")
}

func (r *SpaceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data SpaceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	spaceId, err := uuid.Parse(data.SpaceId.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid space_id", "space_id should be in UUID format")
		return
	}

	organisationId, err := uuid.Parse(data.OrganisationId.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid organisation_id", "organisation_id should be in UUID format")
		return
	}

	res := utils.ExecuteRequest(func() (*numspot.DeleteSpaceResponse, error) {
		return numspotClient.DeleteSpaceWithResponse(ctx, organisationId, spaceId)
	}, http.StatusNoContent, &response.Diagnostics)
	if res == nil {
		return
	}

	response.State.RemoveResource(ctx)
}
