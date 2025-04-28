package vpc

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
	"terraform-provider-numspot/internal/services/tags"
	"terraform-provider-numspot/internal/services/vpc/resource_vpc"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &vpcResource{}
	_ resource.ResourceWithConfigure   = &vpcResource{}
	_ resource.ResourceWithImportState = &vpcResource{}
)

type vpcResource struct {
	provider *client.NumSpotSDK
}

func NewVPCResource() resource.Resource {
	return &vpcResource{}
}

func (r *vpcResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *vpcResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *vpcResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_vpc"
}

func (r *vpcResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_vpc.VpcResourceSchema(ctx)
}

func (r *vpcResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_vpc.VpcModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := vpcTags(ctx, plan.Tags)
	dhcpOptionsSet := plan.DhcpOptionsSetId.ValueString()

	numSpotVPC, err := core.CreateVPC(ctx, r.provider, deserializeCreateVPCRequest(plan), dhcpOptionsSet, tagsValue)
	if err != nil {
		response.Diagnostics.AddError("unable to create vpc", err.Error())
		return
	}

	state := serializeVPC(ctx, numSpotVPC, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *vpcResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_vpc.VpcModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	vpcID := state.Id.ValueString()

	numSpotVPC, err := core.ReadVPC(ctx, r.provider, vpcID)
	if err != nil {
		response.Diagnostics.AddError("unable to read vpc", err.Error())
		return
	}

	newState := serializeVPC(ctx, numSpotVPC, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *vpcResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err         error
		state, plan resource_vpc.VpcModel
		numSpotVPC  *api.Vpc
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	vpcID := state.Id.ValueString()
	planTags := vpcTags(ctx, plan.Tags)
	stateTags := vpcTags(ctx, state.Tags)

	if !plan.Tags.Equal(state.Tags) {
		numSpotVPC, err = core.UpdateVPCTags(ctx, r.provider, vpcID, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update vpc tags", err.Error())
			return
		}
	}

	newState := serializeVPC(ctx, numSpotVPC, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *vpcResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_vpc.VpcModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteVPC(ctx, r.provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete vpc", err.Error())
		return
	}
}

func serializeVPC(ctx context.Context, http *api.Vpc, diags *diag.Diagnostics) resource_vpc.VpcModel {
	var tagsTf types.List

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return resource_vpc.VpcModel{
		DhcpOptionsSetId: types.StringPointerValue(http.DhcpOptionsSetId),
		Id:               types.StringPointerValue(http.Id),
		IpRange:          types.StringPointerValue(http.IpRange),
		State:            types.StringPointerValue(http.State),
		Tenancy:          types.StringPointerValue(http.Tenancy),
		Tags:             tagsTf,
	}
}

func deserializeCreateVPCRequest(tf resource_vpc.VpcModel) api.CreateVpcJSONRequestBody {
	return api.CreateVpcJSONRequestBody{
		IpRange: tf.IpRange.ValueString(),
		Tenancy: tf.Tenancy.ValueStringPointer(),
	}
}

func vpcTags(ctx context.Context, tags types.List) []api.ResourceTag {
	tfTags := make([]resource_vpc.TagsValue, 0, len(tags.Elements()))
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
