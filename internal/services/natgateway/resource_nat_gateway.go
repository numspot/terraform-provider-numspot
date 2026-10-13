package natgateway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/natgateway/resource_nat_gateway"
	"terraform-provider-numspot/internal/services/tags"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &natGatewayResource{}
	_ resource.ResourceWithConfigure   = &natGatewayResource{}
	_ resource.ResourceWithImportState = &natGatewayResource{}
)

// natGatewayResource is the Terraform resource implementation for numspot_nat_gateway.
type natGatewayResource struct {
	services.BaseResource
}

// NewNatGatewayResource returns a new natGatewayResource instance.
func NewNatGatewayResource() resource.Resource {
	return &natGatewayResource{}
}

// Metadata sets the resource type name.
func (r *natGatewayResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_nat_gateway"
}

// Schema defines the schema for the resource.
func (r *natGatewayResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_nat_gateway.NatGatewayResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *natGatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_nat_gateway.NatGatewayModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := natgatewayTags(ctx, plan.Tags)
	numSpotNatGateway, err := core.CreateNatGateway(ctx, r.Provider, deserializeCreateNatGatewayRequest(&plan), tagsValue)
	if err != nil {
		response.Diagnostics.AddError("unable to create nat_gateway", err.Error())
		return
	}

	state := serializeNatGateway(ctx, numSpotNatGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *natGatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_nat_gateway.NatGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numSpotNatGateway, err := core.ReadNatGateway(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read nat_gateway", err.Error())
		return
	}

	newState := serializeNatGateway(ctx, numSpotNatGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *natGatewayResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err               error
		state, plan       resource_nat_gateway.NatGatewayModel
		numSpotNatGateway *api.NatGateway
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	natgatewayID := state.Id.ValueString()
	planTags := natgatewayTags(ctx, plan.Tags)
	stateTags := natgatewayTags(ctx, state.Tags)

	if !plan.Tags.Equal(state.Tags) {
		numSpotNatGateway, err = core.UpdateNatGatewayTags(ctx, r.Provider, natgatewayID, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update nat_gateway tags", err.Error())
			return
		}
	}

	newState := serializeNatGateway(ctx, numSpotNatGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *natGatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_nat_gateway.NatGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteNatGateway(ctx, r.Provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete nat_gateway", err.Error())
		return
	}
}

func serializeNatGateway(ctx context.Context, http *api.NatGateway, diags *diag.Diagnostics) resource_nat_gateway.NatGatewayModel {
	var tagsTf types.Set
	if http == nil {
		diags.AddError("nil NatGateway response", "the API returned an empty NatGateway")
		return resource_nat_gateway.NatGatewayModel{}
	}
	if http.Tags != nil {
		tagsTf = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return resource_nat_gateway.NatGatewayModel{
		Id:         types.StringPointerValue(http.Id),
		PublicIpId: firstPublicIpId(http.PublicIps),
		PublicIps:  utils.GenericListToTfListValue(ctx, serializePublicIp, *http.PublicIps, diags),
		State:      types.StringPointerValue(http.State),
		SubnetId:   types.StringPointerValue(http.SubnetId),
		VpcId:      types.StringPointerValue(http.VpcId),
		Tags:       tagsTf,
	}
}

func deserializeCreateNatGatewayRequest(tf *resource_nat_gateway.NatGatewayModel) *api.CreateNatGatewayJSONRequestBody {
	return &api.CreateNatGatewayJSONRequestBody{
		PublicIpId: tf.PublicIpId.ValueString(),
		SubnetId:   tf.SubnetId.ValueString(),
	}
}

func natgatewayTags(ctx context.Context, tagsSet types.Set) []api.ResourceTag {
	tfTags := make([]resource_nat_gateway.TagsValue, 0, len(tagsSet.Elements()))
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

func serializePublicIp(ctx context.Context, item *api.PublicIpLight, diags *diag.Diagnostics) resource_nat_gateway.PublicIpsValue {
	value, d := resource_nat_gateway.NewPublicIpsValue(resource_nat_gateway.PublicIpsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"public_ip":    types.StringPointerValue(item.PublicIp),
		"public_ip_id": types.StringPointerValue(item.PublicIpId),
	})
	diags.Append(d...)
	return value
}
