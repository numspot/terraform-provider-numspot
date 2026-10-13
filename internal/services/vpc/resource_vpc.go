package vpc

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

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

// vpcResource is the Terraform resource implementation for numspot_vpc.
type vpcResource struct {
	services.BaseResource
}

// NewVPCResource returns a new vpcResource instance.
func NewVPCResource() resource.Resource {
	return &vpcResource{}
}

// Metadata sets the resource type name.
func (r *vpcResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_vpc"
}

// Schema defines the schema for the resource.
func (r *vpcResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_vpc.VpcResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *vpcResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_vpc.VpcModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := vpcTags(ctx, plan.Tags)
	numSpotVPC, err := core.CreateVPC(ctx, r.Provider, deserializeCreateVPCRequest(&plan), plan.DhcpOptionsSetId.ValueString(), tagsValue)
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

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *vpcResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_vpc.VpcModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numSpotVPC, err := core.ReadVPC(ctx, r.Provider, state.Id.ValueString())
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

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
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
		numSpotVPC, err = core.UpdateVPCTags(ctx, r.Provider, vpcID, stateTags, planTags)
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

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *vpcResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_vpc.VpcModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteVPC(ctx, r.Provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete vpc", err.Error())
		return
	}
}

func serializeVPC(ctx context.Context, http *api.Vpc, diags *diag.Diagnostics) resource_vpc.VpcModel {
	var tagsTf types.Set
	if http == nil {
		diags.AddError("nil Vpc response", "the API returned an empty Vpc")
		return resource_vpc.VpcModel{}
	}
	if http.Tags != nil {
		tagsTf = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
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

func deserializeCreateVPCRequest(tf *resource_vpc.VpcModel) *api.CreateVpcJSONRequestBody {
	return &api.CreateVpcJSONRequestBody{
		IpRange: tf.IpRange.ValueString(),
		Tenancy: tf.Tenancy.ValueStringPointer(),
	}
}

func vpcTags(ctx context.Context, tagsSet types.Set) []api.ResourceTag {
	tfTags := make([]resource_vpc.TagsValue, 0, len(tagsSet.Elements()))
	tagsSet.ElementsAs(ctx, &tfTags, false)

	apiTags := make([]api.ResourceTag, 0, len(tfTags))
	for _, tfTag := range tfTags {
		apiTags = append(apiTags, api.ResourceTag{
			Key:   tfTag.Key.ValueString(),
			Value: tfTag.Value.ValueString(),
		})
	}

	return apiTags
}
