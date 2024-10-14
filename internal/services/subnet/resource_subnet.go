package subnet

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &SubnetResource{}
	_ resource.ResourceWithConfigure   = &SubnetResource{}
	_ resource.ResourceWithImportState = &SubnetResource{}
)

type SubnetResource struct {
	provider *client.NumSpotSDK
}

func NewSubnetResource() resource.Resource {
	return &SubnetResource{}
}

func (r *SubnetResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *SubnetResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *SubnetResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_subnet"
}

func (r *SubnetResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = SubnetResourceSchema(ctx)
}

func (r *SubnetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan SubnetModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := tags.TfTagsToApiTags(ctx, plan.Tags)

	numSpotSubnet, err := core.CreateSubnet(ctx, r.provider, deserializeCreateSubnet(plan), plan.MapPublicIpOnLaunch.ValueBool(), tagsValue)
	if err != nil {
		response.Diagnostics.AddError("unable to create subnet", err.Error())
		return
	}

	state, diags := serializeSubnet(ctx, numSpotSubnet)
	if diags.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *SubnetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state SubnetModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	subnetID := state.Id.ValueString()

	numSpotSubnet, err := core.ReadSubnet(ctx, r.provider, subnetID)
	if err != nil {
		response.Diagnostics.AddError("unable to read subnet", err.Error())
		return
	}

	newState, diags := serializeSubnet(ctx, numSpotSubnet)
	if diags.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *SubnetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err           error
		numSpotSubnet *numspot.Subnet
		plan, state   SubnetModel
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	subnetID := state.Id.ValueString()
	mapPublicIPOnLaunch := plan.MapPublicIpOnLaunch.ValueBool()
	planTags := tags.TfTagsToApiTags(ctx, plan.Tags)
	stateTags := tags.TfTagsToApiTags(ctx, state.Tags)

	if !plan.MapPublicIpOnLaunch.Equal(state.MapPublicIpOnLaunch) {
		if numSpotSubnet, err = core.UpdateSubnetAttributes(ctx, r.provider, subnetID, mapPublicIPOnLaunch); err != nil {
			response.Diagnostics.AddError("unable to update subnet attributes", err.Error())
			return
		}
	}

	if !plan.Tags.Equal(state.Tags) {
		if numSpotSubnet, err = core.UpdateSubnetTags(ctx, r.provider, subnetID, stateTags, planTags); err != nil {
			response.Diagnostics.AddError("unable to update subnet tags", err.Error())
		}
	}

	newState, diags := serializeSubnet(ctx, numSpotSubnet)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *SubnetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state SubnetModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	subnetID := state.Id.ValueString()

	if err := core.DeleteSubnet(ctx, r.provider, subnetID); err != nil {
		response.Diagnostics.AddError("unable to delete subnet", err.Error())
	}
}

func deserializeCreateSubnet(tf SubnetModel) numspot.CreateSubnetJSONRequestBody {
	return numspot.CreateSubnetJSONRequestBody{
		IpRange:              tf.IpRange.ValueString(),
		VpcId:                tf.VpcId.ValueString(),
		AvailabilityZoneName: tf.AvailabilityZoneName.ValueStringPointer(),
	}
}

func serializeSubnet(ctx context.Context, http *numspot.Subnet) (*SubnetModel, diag.Diagnostics) {
	var (
		tagsList types.List
		diags    diag.Diagnostics
	)
	if http.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, &diags)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &SubnetModel{
		AvailableIpsCount:    utils.FromIntPtrToTfInt64(http.AvailableIpsCount),
		Id:                   types.StringPointerValue(http.Id),
		IpRange:              types.StringPointerValue(http.IpRange),
		MapPublicIpOnLaunch:  types.BoolPointerValue(http.MapPublicIpOnLaunch),
		VpcId:                types.StringPointerValue(http.VpcId),
		State:                types.StringPointerValue(http.State),
		AvailabilityZoneName: types.StringPointerValue(http.AvailabilityZoneName),
		Tags:                 tagsList,
	}, nil
}
