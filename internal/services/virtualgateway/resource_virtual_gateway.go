package virtualgateway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud-sdk/numspot-sdk-go/pkg/numspot"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/client"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/core"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

type Resource struct {
	provider *client.NumSpotSDK
}

func NewVirtualGatewayResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	numSpotClient, ok := request.ProviderData.(*client.NumSpotSDK)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.provider = numSpotClient
}

func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *Resource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_virtual_gateway"
}

func (r *Resource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = VirtualGatewayResourceSchema(ctx)
}

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan VirtualGatewayModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := tags.TfTagsToApiTags(ctx, plan.Tags)
	vpcId := plan.VpcId.ValueString()
	virtualGateway, err := core.CreateVirtualGateway(ctx, r.provider, deserializeCreateVirtualGateway(plan), vpcId, tagsValue)
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

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state *VirtualGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	virtualGatewayID := state.Id.ValueString()

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

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		state, plan           VirtualGatewayModel
		numSpotVirtualGateway *numspot.VirtualGateway
		err                   error
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	clientGatewayID := state.Id.ValueString()
	planTags := tags.TfTagsToApiTags(ctx, plan.Tags)
	stateTags := tags.TfTagsToApiTags(ctx, state.Tags)

	if !state.Tags.Equal(plan.Tags) {
		numSpotVirtualGateway, err = core.UpdateVirtualGatewayTags(ctx, r.provider, stateTags, planTags, clientGatewayID)
		if err != nil {
			response.Diagnostics.AddError("unable to update virtual gateway tags", err.Error())
			return
		}
	}

	state = *serializeVirtualGateway(ctx, numSpotVirtualGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state VirtualGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	virtualGatewayID := state.Id.ValueString()
	vpcId := state.VpcId.ValueString()

	err := core.DeleteVirtualGateway(ctx, r.provider, virtualGatewayID, vpcId)
	if err != nil {
		response.Diagnostics.AddError("unable to delete virtual gateway", err.Error())
		return
	}
}

func deserializeCreateVirtualGateway(tf VirtualGatewayModel) numspot.CreateVirtualGatewayJSONRequestBody {
	return numspot.CreateVirtualGatewayJSONRequestBody{
		ConnectionType: tf.ConnectionType.ValueString(),
	}
}

func serializeVirtualGateway(ctx context.Context, http *numspot.VirtualGateway, diags *diag.Diagnostics) *VirtualGatewayModel {
	var tagsTf, vpcToVirtualGatewayLinksTf types.List

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	if http.VpcToVirtualGatewayLinks != nil {
		vpcToVirtualGatewayLinksTf = utils.GenericListToTfListValue(ctx, serializeVpcToVirtualGatewayLinks, *http.VpcToVirtualGatewayLinks, diags)
		if diags.HasError() {
			return nil
		}
	}

	vpcId := getVpcId(http.VpcToVirtualGatewayLinks)

	return &VirtualGatewayModel{
		ConnectionType:           types.StringPointerValue(http.ConnectionType),
		Id:                       types.StringPointerValue(http.Id),
		VpcToVirtualGatewayLinks: vpcToVirtualGatewayLinksTf,
		VpcId:                    types.StringPointerValue(vpcId),
		State:                    types.StringPointerValue(http.State),
		Tags:                     tagsTf,
	}
}

func serializeVpcToVirtualGatewayLinks(ctx context.Context, http numspot.VpcToVirtualGatewayLink, diags *diag.Diagnostics) VpcToVirtualGatewayLinksValue {
	value, diagnostics := NewVpcToVirtualGatewayLinksValue(VpcToVirtualGatewayLinksValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"state":  types.StringPointerValue(http.State),
			"vpc_id": types.StringPointerValue(http.VpcId),
		},
	)
	diags.Append(diagnostics...)
	return value
}

/*
The Linking of Virtual Gateway with VPC is weird on Outscale side :
A Virtual Gateway can be linked to a single VPC, but vpcToVirtualGatewayLinks is an array of VPCs
This array contains all the VPC that has been linked to the Virtual Gateway (a given VPC can appear multiple time)
VPCs that have been unlinked have the state "detached", and the linked VPC (if any), have state "attached"

This function retrieve the first (single) vpcId that has state attached, if any
*/
func getVpcId(vpcToVirtualGatewayLinks *[]numspot.VpcToVirtualGatewayLink) *string {
	var vpcId *string

	if vpcToVirtualGatewayLinks != nil {
		vpcToVirtualGatewayLinksValue := *vpcToVirtualGatewayLinks

		for _, vpc := range vpcToVirtualGatewayLinksValue {
			if vpc.State != nil && vpc.VpcId != nil && *vpc.State == "attached" {
				vpcId = vpc.VpcId
				break
			}
		}
	}

	return vpcId
}
