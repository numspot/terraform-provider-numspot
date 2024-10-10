package vpc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &VpcResource{}
	_ resource.ResourceWithConfigure   = &VpcResource{}
	_ resource.ResourceWithImportState = &VpcResource{}
)

type VpcResource struct {
	provider *client.NumSpotSDK
}

func NewNetResource() resource.Resource {
	return &VpcResource{}
}

func (r *VpcResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *VpcResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *VpcResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_vpc"
}

func (r *VpcResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = VpcResourceSchema(ctx)
}

func (r *VpcResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data VpcModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	// Retries create until request response is OK
	res, err := utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		NetFromTfToCreateRequest(&data),
		numspotClient.CreateVpcWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VPC", err.Error())
		return
	}

	// Handle tags
	createdId := *res.JSON201.Id
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, numspotClient, r.provider.SpaceID, &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// if dhcp_options_set_id is set, we need to update the Vpc as this attribute can be set on Update only and not on Create
	if !data.DhcpOptionsSetId.IsNull() && !data.DhcpOptionsSetId.IsUnknown() {
		updatedRes := utils.ExecuteRequest(func() (*numspot.UpdateVpcResponse, error) {
			body := VpcFromTfToUpdaterequest(ctx, &data)
			return numspotClient.UpdateVpcWithResponse(ctx, r.provider.SpaceID, createdId, body)
		}, http.StatusOK, &response.Diagnostics)

		if updatedRes == nil || response.Diagnostics.HasError() {
			return
		}
	}
	readRes, err := utils.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.SpaceID,
		[]string{"pending"},
		[]string{"available"},
		numspotClient.ReadVpcsByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Net", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", createdId, err))
		return
	}

	vpc, ok := readRes.(*numspot.Vpc)
	if !ok {
		response.Diagnostics.AddError("Failed to read VPC", "object conversion error")
		return
	}

	tf := NetFromHttpToTf(ctx, vpc, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *VpcResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data VpcModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	res := utils.ExecuteRequest(func() (*numspot.ReadVpcsByIdResponse, error) {
		return numspotClient.ReadVpcsByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	// TODO: read Nets returns tags in response, do not need to relist tags
	tf := NetFromHttpToTf(ctx, res.JSON200, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *VpcResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		state VpcModel
		plan  VpcModel
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	vpcId := state.Id.ValueString()

	if !state.Tags.Equal(plan.Tags) {
		tags.UpdateTags(
			ctx,
			state.Tags,
			plan.Tags,
			&response.Diagnostics,
			numspotClient,
			r.provider.SpaceID,
			vpcId,
		)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// Update Vpc
	updatedRes := utils.ExecuteRequest(func() (*numspot.UpdateVpcResponse, error) {
		body := VpcFromTfToUpdaterequest(ctx, &plan)
		return numspotClient.UpdateVpcWithResponse(ctx, r.provider.SpaceID, vpcId, body)
	}, http.StatusOK, &response.Diagnostics)

	if updatedRes == nil || response.Diagnostics.HasError() {
		return
	}

	// Read resource
	res := utils.ExecuteRequest(func() (*numspot.ReadVpcsByIdResponse, error) {
		return numspotClient.ReadVpcsByIdWithResponse(ctx, r.provider.SpaceID, state.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := NetFromHttpToTf(ctx, res.JSON200, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *VpcResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data VpcModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), numspotClient.DeleteVpcWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete VPC", err.Error())
		return
	}

	response.State.RemoveResource(ctx)
}
