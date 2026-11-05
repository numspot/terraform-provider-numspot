package subnet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/subnet/resource_subnet"
	"github.com/numspot/terraform-provider-numspot/internal/services/tags"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &subnetResource{}
	_ resource.ResourceWithConfigure   = &subnetResource{}
	_ resource.ResourceWithImportState = &subnetResource{}
)

// subnetResource is the Terraform resource implementation for numspot_subnet.
type subnetResource struct {
	services.BaseResource
}

// NewSubnetResource returns a new subnetResource instance.
func NewSubnetResource() resource.Resource {
	return &subnetResource{}
}

// Metadata sets the resource type name.
func (r *subnetResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_subnet"
}

// Schema defines the schema for the resource.
func (r *subnetResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_subnet.SubnetResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *subnetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_subnet.SubnetModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := subnetTags(ctx, plan.Tags)
	numspotSubnet, err := core.CreateSubnet(ctx, r.Provider, deserializeCreateSubnetRequest(&plan), plan.MapPublicIpOnLaunch.ValueBool(), tagsValue)
	if err != nil {
		response.Diagnostics.AddError("unable to create subnet", err.Error())
		return
	}

	state := serializeSubnet(ctx, numspotSubnet, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *subnetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_subnet.SubnetModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotSubnet, err := core.ReadSubnet(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read subnet", err.Error())
		return
	}

	newState := serializeSubnet(ctx, numspotSubnet, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *subnetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err           error
		state, plan   resource_subnet.SubnetModel
		numspotSubnet *api.Subnet
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
	planTags := subnetTags(ctx, plan.Tags)
	stateTags := subnetTags(ctx, state.Tags)

	if !plan.MapPublicIpOnLaunch.Equal(state.MapPublicIpOnLaunch) {
		numspotSubnet, err = core.UpdateSubnetAttributes(ctx, r.Provider, state.Id.ValueString(), plan.MapPublicIpOnLaunch.ValueBool())
		if err != nil {
			response.Diagnostics.AddError("unable to update subnet", err.Error())
			return
		}
	}

	if !plan.Tags.Equal(state.Tags) {
		numspotSubnet, err = core.UpdateSubnetTags(ctx, r.Provider, subnetID, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update subnet tags", err.Error())
			return
		}
	}

	newState := serializeSubnet(ctx, numspotSubnet, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *subnetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_subnet.SubnetModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteSubnet(ctx, r.Provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete subnet", err.Error())
		return
	}
}

func serializeSubnet(ctx context.Context, http *api.Subnet, diags *diag.Diagnostics) resource_subnet.SubnetModel {
	var tagsTf types.Set
	if http == nil {
		diags.AddError("nil Subnet response", "the API returned an empty Subnet")
		return resource_subnet.SubnetModel{}
	}
	if http.Tags != nil {
		tagsTf = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return resource_subnet.SubnetModel{
		AvailabilityZoneName: types.StringValue(utils.ConvertAzNamePtrToString(http.AvailabilityZoneName)),
		AvailableIpsCount:    utils.FromIntPtrToTfInt64(http.AvailableIpsCount),
		Id:                   types.StringPointerValue(http.Id),
		IpRange:              types.StringPointerValue(http.IpRange),
		MapPublicIpOnLaunch:  types.BoolPointerValue(http.MapPublicIpOnLaunch),
		State:                types.StringPointerValue(http.State),
		VpcId:                types.StringPointerValue(http.VpcId),
		Tags:                 tagsTf,
	}
}

func deserializeCreateSubnetRequest(tf *resource_subnet.SubnetModel) *api.CreateSubnetJSONRequestBody {
	return &api.CreateSubnetJSONRequestBody{
		AvailabilityZoneName: utils.FromTfStringToAzNamePtr(tf.AvailabilityZoneName),
		IpRange:              tf.IpRange.ValueString(),
		VpcId:                tf.VpcId.ValueString(),
	}
}

func subnetTags(ctx context.Context, tagsSet types.Set) []api.ResourceTag {
	tfTags := make([]resource_subnet.TagsValue, 0, len(tagsSet.Elements()))
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
