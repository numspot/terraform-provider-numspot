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

	createSubnetPayload := deserializeCreateSubnet(plan)
	tags := tags.TfTagsToApiTags(ctx, plan.Tags)
	res, err := core.CreateSubnet(ctx, r.provider, createSubnetPayload, plan.MapPublicIpOnLaunch.ValueBool(), tags)
	if err != nil {
		response.Diagnostics.AddError("Failed to create subnet", err.Error())
		return
	}

	tf, diags := serializeSubnet(ctx, res)
	if diags.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SubnetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data SubnetModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	res, err := core.ReadSubnet(ctx, r.provider, data.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Failed to read subnet", err.Error())
		return
	}

	tf, diags := serializeSubnet(ctx, res)
	if diags.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SubnetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		plan, state SubnetModel
		res         *numspot.Subnet
		err         error
		updated     bool
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !state.Tags.Equal(plan.Tags) {
		updated = true
		planTags := tags.TfTagsToApiTags(ctx, plan.Tags)
		stateTags := tags.TfTagsToApiTags(ctx, state.Tags)
		if res, err = core.UpdateSubnetTags(ctx, r.provider, state.Id.ValueString(), stateTags, planTags); err != nil {
			response.Diagnostics.AddError("Failed to update subnet", fmt.Sprintf("failed to update tags: %s", err.Error()))
		}
	}

	if !utils.IsTfValueNull(plan.MapPublicIpOnLaunch) {
		updated = true
		if res, err = core.UpdateSubnetAttributes(
			ctx,
			r.provider,
			state.Id.ValueString(),
			plan.MapPublicIpOnLaunch.ValueBool(),
		); err != nil {
			response.Diagnostics.AddError("Failed to update subnet", fmt.Sprintf("failed to update MapPublicIPOnLaunch: %s", err.Error()))
			return
		}
	}

	if updated {
		tf, diags := serializeSubnet(ctx, res)
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}

		response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
	}
}

func (r *SubnetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data SubnetModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteSubnet(ctx, r.provider, data.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("Failed to delete subnet", err.Error())
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
