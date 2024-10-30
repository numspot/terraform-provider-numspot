package natgateway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	_ resource.Resource                = &NatGatewayResource{}
	_ resource.ResourceWithConfigure   = &NatGatewayResource{}
	_ resource.ResourceWithImportState = &NatGatewayResource{}
)

type NatGatewayResource struct {
	provider *client.NumSpotSDK
}

func NewNatGatewayResource() resource.Resource {
	return &NatGatewayResource{}
}

func (r *NatGatewayResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *NatGatewayResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *NatGatewayResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_nat_gateway"
}

func (r *NatGatewayResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = NatGatewayResourceSchema(ctx)
}

func (r *NatGatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan NatGatewayModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := tags.TfTagsToApiTags(ctx, plan.Tags)
	body := deserializeCreateNatGateway(plan)
	if response.Diagnostics.HasError() {
		return
	}

	natGateway, err := core.CreateNatGateway(ctx, r.provider, tagsValue, body)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Nat Gateway", err.Error())
		return
	}

	state := serializeNatGateway(ctx, natGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *NatGatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state NatGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	natGatewayID := state.Id.ValueString()

	numSpotNatGateway, err := core.ReadNatGateway(ctx, r.provider, natGatewayID)
	if err != nil {
		response.Diagnostics.AddError("error while reading nat gateway", err.Error())
		return
	}

	newState := serializeNatGateway(ctx, numSpotNatGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *NatGatewayResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan NatGatewayModel
	var numSpotNatGateway *numspot.NatGateway
	var err error
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	natGatewayID := state.Id.ValueString()
	planTags := tags.TfTagsToApiTags(ctx, plan.Tags)
	stateTags := tags.TfTagsToApiTags(ctx, state.Tags)

	if !state.Tags.Equal(plan.Tags) {
		numSpotNatGateway, err = core.UpdateNatGatewayTags(ctx, r.provider, stateTags, planTags, natGatewayID)
		if err != nil {
			response.Diagnostics.AddError("error while updating tags", err.Error())
			return
		}
	}

	newState := *serializeNatGateway(ctx, numSpotNatGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *NatGatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state NatGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if response.Diagnostics.HasError() {
		return
	}

	natGatewayID := state.Id.ValueString()
	err := core.DeleteNatGateway(ctx, r.provider, natGatewayID)
	if err != nil {
		response.Diagnostics.AddError("error while deleting nat gateway", err.Error())
		return
	}
}

func deserializeCreateNatGateway(tf NatGatewayModel) numspot.CreateNatGatewayJSONRequestBody {
	return numspot.CreateNatGatewayJSONRequestBody{
		PublicIpId: tf.PublicIpId.ValueString(),
		SubnetId:   tf.SubnetId.ValueString(),
	}
}

func serializePublicIp(ctx context.Context, elt numspot.PublicIpLight, diags *diag.Diagnostics) PublicIpsValue {
	value, diagnostics := NewPublicIpsValue(
		PublicIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"public_ip":    types.StringPointerValue(elt.PublicIp),
			"public_ip_id": types.StringPointerValue(elt.PublicIpId),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func serializeNatGateway(ctx context.Context, http *numspot.NatGateway, diags *diag.Diagnostics) *NatGatewayModel {
	var tagsTf types.List

	var publicIp []numspot.PublicIpLight
	if http.PublicIps != nil {
		publicIp = *http.PublicIps
	}
	// Public Ips
	publicIpsTf := utils.GenericListToTfListValue(
		ctx,
		serializePublicIp,
		publicIp,
		diags,
	)
	if diags.HasError() {
		return nil
	}

	// PublicIpId must be the id of the first public io
	var publicIpId *string
	if len(publicIp) > 0 {
		publicIpId = publicIp[0].PublicIpId
	} else {
		publicIpId = nil
	}

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	return &NatGatewayModel{
		Id:         types.StringPointerValue(http.Id),
		PublicIps:  publicIpsTf,
		State:      types.StringPointerValue(http.State),
		SubnetId:   types.StringPointerValue(http.SubnetId),
		VpcId:      types.StringPointerValue(http.VpcId),
		Tags:       tagsTf,
		PublicIpId: types.StringPointerValue(publicIpId),
	}
}
