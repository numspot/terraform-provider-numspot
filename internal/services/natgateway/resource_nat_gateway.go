package natgateway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services/natgateway/resource_nat_gateway"
	"terraform-provider-numspot/internal/services/tags"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

type Resource struct {
	provider *client.NumSpotSDK
}

func NewNatGatewayResource() resource.Resource {
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
	response.TypeName = request.ProviderTypeName + "_nat_gateway"
}

func (r *Resource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_nat_gateway.NatGatewayResourceSchema(ctx)
}

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_nat_gateway.NatGatewayModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := natGatewayTags(ctx, plan.Tags)

	natGateway, err := core.CreateNATGateway(ctx, r.provider, tagsValue, deserializeCreateNATGateway(plan))
	if err != nil {
		response.Diagnostics.AddError("unable to create nat gateway", err.Error())
		return
	}

	state := serializeNATGateway(ctx, natGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_nat_gateway.NatGatewayModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	natGatewayID := state.Id.ValueString()

	numSpotNatGateway, err := core.ReadNATGateway(ctx, r.provider, natGatewayID)
	if err != nil {
		response.Diagnostics.AddError("unable to read nat gateway", err.Error())
		return
	}

	newState := serializeNATGateway(ctx, numSpotNatGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		state, plan       resource_nat_gateway.NatGatewayModel
		numSpotNatGateway *api.NatGateway
		err               error
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	natGatewayID := state.Id.ValueString()
	planTags := natGatewayTags(ctx, plan.Tags)
	stateTags := natGatewayTags(ctx, state.Tags)

	if !state.Tags.Equal(plan.Tags) {
		numSpotNatGateway, err = core.UpdateNATGatewayTags(ctx, r.provider, stateTags, planTags, natGatewayID)
		if err != nil {
			response.Diagnostics.AddError("unable to update nat gateway tags", err.Error())
			return
		}
	}

	newState := serializeNATGateway(ctx, numSpotNatGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_nat_gateway.NatGatewayModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	natGatewayID := state.Id.ValueString()

	if err := core.DeleteNATGateway(ctx, r.provider, natGatewayID); err != nil {
		response.Diagnostics.AddError("unable to delete nat gateway", err.Error())
		return
	}
}

func deserializeCreateNATGateway(tf resource_nat_gateway.NatGatewayModel) api.CreateNatGatewayJSONRequestBody {
	return api.CreateNatGatewayJSONRequestBody{
		PublicIpId: tf.PublicIpId.ValueString(),
		SubnetId:   tf.SubnetId.ValueString(),
	}
}

func serializePublicIp(ctx context.Context, elt api.PublicIpLight, diags *diag.Diagnostics) resource_nat_gateway.PublicIpsValue {
	value, diagnostics := resource_nat_gateway.NewPublicIpsValue(
		resource_nat_gateway.PublicIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"public_ip":    types.StringPointerValue(elt.PublicIp),
			"public_ip_id": types.StringPointerValue(elt.PublicIpId),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func serializeNATGateway(ctx context.Context, http *api.NatGateway, diags *diag.Diagnostics) resource_nat_gateway.NatGatewayModel {
	var tagsTf types.List

	var publicIp []api.PublicIpLight
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
		return resource_nat_gateway.NatGatewayModel{}
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
			return resource_nat_gateway.NatGatewayModel{}
		}
	}

	return resource_nat_gateway.NatGatewayModel{
		Id:         types.StringPointerValue(http.Id),
		PublicIps:  publicIpsTf,
		State:      types.StringPointerValue(http.State),
		SubnetId:   types.StringPointerValue(http.SubnetId),
		VpcId:      types.StringPointerValue(http.VpcId),
		Tags:       tagsTf,
		PublicIpId: types.StringPointerValue(publicIpId),
	}
}

func natGatewayTags(ctx context.Context, tags types.List) []api.ResourceTag {
	tfTags := make([]resource_nat_gateway.TagsValue, 0, len(tags.Elements()))
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
