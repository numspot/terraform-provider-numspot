package virtualgateway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/virtualgateway/resource_virtual_gateway"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &virtualGatewayResource{}
	_ resource.ResourceWithConfigure   = &virtualGatewayResource{}
	_ resource.ResourceWithImportState = &virtualGatewayResource{}
)

// virtualGatewayResource is the Terraform resource implementation for numspot_virtual_gateway.
type virtualGatewayResource struct {
	services.BaseResource
}

// NewVirtualGatewayResource returns a new virtualGatewayResource instance.
func NewVirtualGatewayResource() resource.Resource {
	return &virtualGatewayResource{}
}

// Metadata sets the resource type name.
func (r *virtualGatewayResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_virtual_gateway"
}

// Schema defines the schema for the resource.
func (r *virtualGatewayResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_virtual_gateway.VirtualGatewayResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *virtualGatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_virtual_gateway.VirtualGatewayModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	numspotVirtualGateway, err := core.CreateVirtualGateway(ctx, r.Provider, deserializeCreateVirtualGatewayRequest(&plan), plan.VpcId.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to create virtual_gateway", err.Error())
		return
	}

	state := serializeVirtualGateway(ctx, numspotVirtualGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *virtualGatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_virtual_gateway.VirtualGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotVirtualGateway, err := core.ReadVirtualGateway(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		if utils.IsResourceGone(err) {
			tflog.Warn(ctx, "unable to read virtual_gateway (not found or access forbidden); removing from state")
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("unable to read virtual_gateway", err.Error())
		return
	}

	newState := serializeVirtualGateway(ctx, numspotVirtualGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *virtualGatewayResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *virtualGatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_virtual_gateway.VirtualGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteVirtualGateway(ctx, r.Provider, state.Id.ValueString(), state.VpcId.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete virtual_gateway", err.Error())
		return
	}
}

func serializeVirtualGateway(ctx context.Context, http *api.VirtualGateway, diags *diag.Diagnostics) resource_virtual_gateway.VirtualGatewayModel {
	if http == nil {
		diags.AddError("nil VirtualGateway response", "the API returned an empty VirtualGateway")
		return resource_virtual_gateway.VirtualGatewayModel{}
	}

	return resource_virtual_gateway.VirtualGatewayModel{
		ConnectionType:           types.StringValue(http.ConnectionType),
		Id:                       types.StringValue(http.Id.String()),
		State:                    types.StringValue(http.State),
		VpcId:                    types.StringValue(getVpcId(http.VpcToVirtualGatewayLinks)),
		VpcToVirtualGatewayLinks: utils.GenericListToTfListValue(ctx, serializeVpcToVirtualGatewayLink, http.VpcToVirtualGatewayLinks, diags),
	}
}

func deserializeCreateVirtualGatewayRequest(tf *resource_virtual_gateway.VirtualGatewayModel) *api.CreateVirtualGatewayJSONRequestBody {
	return &api.CreateVirtualGatewayJSONRequestBody{
		ConnectionType: tf.ConnectionType.ValueString(),
	}
}

func serializeVpcToVirtualGatewayLink(
	ctx context.Context,
	item *api.VpcToVirtualGatewayLink,
	diags *diag.Diagnostics,
) resource_virtual_gateway.VpcToVirtualGatewayLinksValue {
	value, d := resource_virtual_gateway.NewVpcToVirtualGatewayLinksValue(
		resource_virtual_gateway.VpcToVirtualGatewayLinksValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"state":  types.StringValue(item.State),
			"vpc_id": types.StringValue(item.VpcId),
		},
	)
	diags.Append(d...)
	return value
}
