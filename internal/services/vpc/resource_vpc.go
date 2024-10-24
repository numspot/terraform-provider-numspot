package vpc

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
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
	var (
		plan          VpcModel
		updatePayload *numspot.UpdateVpcJSONRequestBody
	)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	createPayload := deserializeCreateVPCRequest(&plan)
	tagsValue := tags.TfTagsToApiTags(ctx, plan.Tags)
	if !plan.DhcpOptionsSetId.IsNull() && !plan.DhcpOptionsSetId.IsUnknown() {
		p := deserializeUpdateVPCRequest(ctx, &plan)
		updatePayload = &p
	}
	vpc, err := core.CreateVPC(ctx, r.provider, createPayload, tagsValue, updatePayload)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VPC", err.Error())
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
		deserializeCreateVPCRequest(&plan),
		numspotClient.CreateVpcWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VPC", err.Error())
		return
	}

	// Handle tags
	createdId := *res.JSON201.Id
	if len(plan.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, numspotClient, r.provider.SpaceID, &response.Diagnostics, createdId, plan.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// if dhcp_options_set_id is set, we need to update the Vpc as this attribute can be set on Update only and not on Create
	if !plan.DhcpOptionsSetId.IsNull() && !plan.DhcpOptionsSetId.IsUnknown() {
		updatedRes := utils.ExecuteRequest(func() (*numspot.UpdateVpcResponse, error) {
			body := deserializeUpdateVPCRequest(ctx, &plan)
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

	tf := serializeVPC(ctx, vpc, &response.Diagnostics)
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

	vpc, err := core.ReadVPC(ctx, r.provider, data.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("failed to read VPC", err.Error())
		return
	}

	tf := serializeVPC(ctx, vpc, &response.Diagnostics)
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
		body := deserializeUpdateVPCRequest(ctx, &plan)
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

	tf := serializeVPC(ctx, res.JSON200, &response.Diagnostics)
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

	if err := core.DeleteVPC(ctx, r.provider, data.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("failed to delete VPC", err.Error())
		return
	}
}

func serializeVPC(ctx context.Context, http *numspot.Vpc, diags *diag.Diagnostics) *VpcModel {
	var tagsTf types.List

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return &VpcModel{
		DhcpOptionsSetId: types.StringPointerValue(http.DhcpOptionsSetId),
		Id:               types.StringPointerValue(http.Id),
		IpRange:          types.StringPointerValue(http.IpRange),
		State:            types.StringPointerValue(http.State),
		Tenancy:          types.StringPointerValue(http.Tenancy),
		Tags:             tagsTf,
	}
}

func deserializeUpdateVPCRequest(ctx context.Context, tf *VpcModel) numspot.UpdateVpcJSONRequestBody {
	return numspot.UpdateVpcJSONRequestBody{
		DhcpOptionsSetId: tf.DhcpOptionsSetId.ValueString(),
	}
}

func deserializeCreateVPCRequest(tf *VpcModel) numspot.CreateVpcJSONRequestBody {
	return numspot.CreateVpcJSONRequestBody{
		IpRange: tf.IpRange.ValueString(),
		Tenancy: tf.Tenancy.ValueStringPointer(),
	}
}
