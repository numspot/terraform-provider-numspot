package subnet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/subnet/resource_subnet"
	"terraform-provider-numspot/internal/services/tags"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &subnetResource{}
	_ resource.ResourceWithConfigure   = &subnetResource{}
	_ resource.ResourceWithImportState = &subnetResource{}
)

type subnetResource struct {
	provider *client.NumSpotSDK
}

func NewSubnetResource() resource.Resource {
	return &subnetResource{}
}

func (r *subnetResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *subnetResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *subnetResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_subnet"
}

func (r *subnetResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_subnet.SubnetResourceSchema(ctx)
}

func (r *subnetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_subnet.SubnetModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := subnetTags(ctx, plan.Tags)
	mapPublicIP := plan.MapPublicIpOnLaunch.ValueBool()

	numSpotSubnet, err := core.CreateSubnet(ctx, r.provider, deserializeCreateSubnet(plan), mapPublicIP, tagsValue)
	if err != nil {
		response.Diagnostics.AddError("unable to create subnet", err.Error())
		return
	}

	state := serializeSubnet(ctx, numSpotSubnet, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *subnetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_subnet.SubnetModel
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

	newState := serializeSubnet(ctx, numSpotSubnet, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *subnetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err           error
		numSpotSubnet *api.Subnet
		plan, state   resource_subnet.SubnetModel
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
	planTags := subnetTags(ctx, plan.Tags)
	stateTags := subnetTags(ctx, state.Tags)

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

	newState := serializeSubnet(ctx, numSpotSubnet, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *subnetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_subnet.SubnetModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	subnetID := state.Id.ValueString()

	if err := core.DeleteSubnet(ctx, r.provider, subnetID); err != nil {
		response.Diagnostics.AddError("unable to delete subnet", err.Error())
	}
}

func deserializeCreateSubnet(tf resource_subnet.SubnetModel) api.CreateSubnetJSONRequestBody {
	return api.CreateSubnetJSONRequestBody{
		IpRange:              tf.IpRange.ValueString(),
		VpcId:                tf.VpcId.ValueString(),
		AvailabilityZoneName: utils.FromTfStringToAzNamePtr(tf.AvailabilityZoneName),
	}
}

func serializeSubnet(ctx context.Context, http *api.Subnet, diags *diag.Diagnostics) *resource_subnet.SubnetModel {
	var tagsList types.Set
	if http.Tags != nil {
		tagsList = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	return &resource_subnet.SubnetModel{
		AvailableIpsCount:    utils.FromIntPtrToTfInt64(http.AvailableIpsCount),
		Id:                   types.StringPointerValue(http.Id),
		IpRange:              types.StringPointerValue(http.IpRange),
		MapPublicIpOnLaunch:  types.BoolPointerValue(http.MapPublicIpOnLaunch),
		VpcId:                types.StringPointerValue(http.VpcId),
		State:                types.StringPointerValue(http.State),
		AvailabilityZoneName: types.StringValue(utils.ConvertAzNamePtrToString(http.AvailabilityZoneName)),
		Tags:                 tagsList,
	}
}

func subnetTags(ctx context.Context, tags types.Set) []api.ResourceTag {
	tfTags := make([]resource_subnet.TagsValue, 0, len(tags.Elements()))
	tags.ElementsAs(ctx, &tfTags, false)

	apiTags := make([]api.ResourceTag, 0, len(tfTags))
	for _, tfTag := range tfTags {
		apiTags = append(apiTags, api.ResourceTag{
			Key:   tfTag.Key.ValueString(),
			Value: tfTag.Value.ValueString(),
		})
	}

	return apiTags
}
