// Package vpc provides the implementation of the VPC (Virtual Private Cloud) resource
// for the NumSpot provider. It handles the creation, reading, updating, and deletion
// of VPCs in NumSpot, including managing VPC attributes such as IP ranges, tenancy,
// DHCP options, and tags.
package vpc

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services/tags"
	"terraform-provider-numspot/internal/services/vpc/resource_vpc"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

// Resource represents the VPC resource implementation.
// It implements the Terraform resource.Resource interface and provides
// CRUD operations for VPCs in NumSpot.
type Resource struct {
	provider *client.NumSpotSDK
}

// NewNetResource creates a new instance of the VPC resource.
// This is the factory function used by the provider to create new VPC resource instances.
func NewNetResource() resource.Resource {
	return &Resource{}
}

// Configure implements the resource.ResourceWithConfigure interface.
// It configures the resource with the provider's SDK client.
func (r *Resource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

// ImportState implements the resource.ResourceWithImportState interface.
// It allows importing existing VPCs into Terraform state using their ID.
func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

// Metadata implements the resource.Resource interface.
// It sets the resource type name for the VPC resource.
func (r *Resource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_vpc"
}

// Schema implements the resource.Resource interface.
// It defines the schema for the VPC resource, including all its attributes
// and their types.
func (r *Resource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_vpc.VpcResourceSchema(ctx)
}

// Create implements the resource.Resource interface.
// It creates a new VPC in NumSpot with the specified configuration.
// The function handles:
// - Setting up the VPC with the specified IP range and tenancy
// - Configuring DHCP options if specified
// - Applying any provided tags
func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
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

// Read implements the resource.Resource interface.
// It reads the current state of an existing VPC from NumSpot.
// The function retrieves all VPC attributes including its configuration,
// DHCP options, and tags.
func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
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

// Update implements the resource.Resource interface.
// It updates an existing VPC in NumSpot.
// Currently, it only supports updating tags, as other VPC attributes
// are immutable after creation.
func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
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

// Delete implements the resource.Resource interface.
// It deletes an existing VPC from NumSpot.
// The deletion is asynchronous and the function will wait for
// the deletion to complete before returning.
func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
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

// serializeVPC converts a NumSpot API VPC response into the Terraform
// resource model. It handles all VPC attributes including DHCP options,
// IP range, state, tenancy, and tags.
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

// deserializeCreateVPCRequest converts the Terraform resource model
// into the NumSpot API request format for VPC creation.
// It handles the required fields for VPC creation including IP range
// and tenancy configuration.
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
