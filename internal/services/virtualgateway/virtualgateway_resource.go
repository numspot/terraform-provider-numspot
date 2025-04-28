package virtualgateway

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/virtualgateway/resource_virtual_gateway"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &virtualGatewayResource{}
	_ resource.ResourceWithConfigure   = &virtualGatewayResource{}
	_ resource.ResourceWithImportState = &virtualGatewayResource{}
)

type virtualGatewayResource struct {
	provider *client.NumSpotSDK
}

func NewVirtualGatewayResource() resource.Resource {
	return &virtualGatewayResource{}
}

func (r *virtualGatewayResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *virtualGatewayResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *virtualGatewayResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_virtual_gateway"
}

func (r *virtualGatewayResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_virtual_gateway.VirtualGatewayResourceSchema(ctx)
}

func (r *virtualGatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_virtual_gateway.VirtualGatewayModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	vpcId := plan.VpcId.ValueString()
	virtualGateway, err := core.CreateVirtualGateway(ctx, r.provider, deserializeCreateVirtualGateway(plan), vpcId)
	if err != nil {
		response.Diagnostics.AddError("unable to create virtual gateway", err.Error())
		return
	}

	state := serializeVirtualGateway(ctx, virtualGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *virtualGatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state *resource_virtual_gateway.VirtualGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	virtualGatewayID, err := uuid.Parse(state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to parse id from state", err.Error())
		return
	}

	numSpotVirtualGateway, err := core.ReadVirtualGateway(ctx, r.provider, virtualGatewayID)
	if err != nil {
		response.Diagnostics.AddError("unable to read virtual gateway", err.Error())
		return
	}

	state = serializeVirtualGateway(ctx, numSpotVirtualGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (r *virtualGatewayResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
}

func (r *virtualGatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_virtual_gateway.VirtualGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	virtualGatewayID, err := uuid.Parse(state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to parse id from state", err.Error())
		return
	}
	vpcId := state.VpcId.ValueString()

	err = core.DeleteVirtualGateway(ctx, r.provider, virtualGatewayID, vpcId)
	if err != nil {
		response.Diagnostics.AddError("unable to delete virtual gateway", err.Error())
		return
	}
}

func deserializeCreateVirtualGateway(tf resource_virtual_gateway.VirtualGatewayModel) api.CreateVirtualGatewayJSONRequestBody {
	return api.CreateVirtualGatewayJSONRequestBody{
		ConnectionType: tf.ConnectionType.ValueString(),
	}
}

func serializeVirtualGateway(ctx context.Context, http *api.VirtualGateway, diags *diag.Diagnostics) *resource_virtual_gateway.VirtualGatewayModel {
	var vpcToVirtualGatewayLinksTf types.List

	if http.VpcToVirtualGatewayLinks != nil {
		vpcToVirtualGatewayLinksTf = utils.GenericListToTfListValue(ctx, serializeVpcToVirtualGatewayLinks, http.VpcToVirtualGatewayLinks, diags)
		if diags.HasError() {
			return nil
		}
	}

	vpcId := getVpcId(http.VpcToVirtualGatewayLinks)

	return &resource_virtual_gateway.VirtualGatewayModel{
		ConnectionType:           types.StringValue(http.ConnectionType),
		Id:                       types.StringValue(http.Id.String()),
		VpcToVirtualGatewayLinks: vpcToVirtualGatewayLinksTf,
		VpcId:                    types.StringValue(vpcId),
		State:                    types.StringValue(http.State),
	}
}

func serializeVpcToVirtualGatewayLinks(ctx context.Context, http api.VpcToVirtualGatewayLink, diags *diag.Diagnostics) resource_virtual_gateway.VpcToVirtualGatewayLinksValue {
	value, diagnostics := resource_virtual_gateway.NewVpcToVirtualGatewayLinksValue(resource_virtual_gateway.VpcToVirtualGatewayLinksValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"state":  types.StringValue(http.State),
			"vpc_id": types.StringValue(http.VpcId),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func getVpcId(vpcToVirtualGatewayLinks []api.VpcToVirtualGatewayLink) string {
	var vpcId string

	if vpcToVirtualGatewayLinks != nil {
		vpcToVirtualGatewayLinksValue := vpcToVirtualGatewayLinks

		for _, vpc := range vpcToVirtualGatewayLinksValue {
			if vpc.VpcId != "" && vpc.State == "attached" {
				vpcId = vpc.VpcId
				break
			}
		}
	}

	return vpcId
}
