package vpcpeering

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud-sdk/numspot-sdk-go/pkg/numspot"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/client"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/core"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource              = &Resource{}
	_ resource.ResourceWithConfigure = &Resource{}
)

type Resource struct {
	provider *client.NumSpotSDK
}

func NewVpcPeeringResource() resource.Resource {
	return &Resource{}
}

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

func (r *Resource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_vpc_peering"
}

func (r *Resource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = VpcPeeringResourceSchema(ctx)
}

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan VpcPeeringModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	payload := deserializeVpcPeering(plan)
	tagsValue := tags.TfTagsToApiTags(ctx, plan.Tags)
	vpcPeering, err := core.CreateVPCPeering(ctx, r.provider, payload, tagsValue)
	if err != nil {
		response.Diagnostics.AddError("failed to create VPC peering", err.Error())
		return
	}

	tf := serializeVpcPeering(ctx, vpcPeering, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data VpcPeeringModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	vpcPeering, err := core.ReadVPCPeering(ctx, r.provider, data.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("failed to read VPC peering", err.Error())
		return
	}

	tf := serializeVpcPeering(ctx, vpcPeering, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan VpcPeeringModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if state.Tags.Equal(plan.Tags) { // Nothing to do here
		return
	}

	stateTags := tags.TfTagsToApiTags(ctx, state.Tags)
	planTags := tags.TfTagsToApiTags(ctx, plan.Tags)
	vpcPeering, err := core.UpdateVPCPeeringTags(ctx, r.provider, state.Id.ValueString(), stateTags, planTags)
	if err != nil {
		response.Diagnostics.AddError("failed to update VPC peering", err.Error())
		return
	}

	tf := serializeVpcPeering(ctx, vpcPeering, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data VpcPeeringModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	if err := core.DeleteVPCPeering(ctx, r.provider, data.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("failed to delete VPC peering", err.Error())
	}
}

func serializeVpcPeering(ctx context.Context, http *numspot.VpcPeering, diags *diag.Diagnostics) *VpcPeeringModel {
	// In the event that the creation of VPC peering fails, the error message might be found in
	// the "state" field. If the state's name is "failed", then the error message will be contained
	// in the state's message. We must address this particular scenario.
	var tagsTf types.List

	vpcPeeringStateHttp := http.State

	if vpcPeeringStateHttp != nil {
		// message := vpcPeeringStateHttp.Message
		name := vpcPeeringStateHttp.Name

		if name != nil && *name == "failed" {
			var errorMessage string
			//if message != nil {
			//	errorMessage = *message
			//}
			diags.AddError("Failed to create vpc peering", errorMessage)
			return nil
		}
	}

	vpcPeeringState := serializeVpcPeeringState(ctx, vpcPeeringStateHttp, diags)
	accepterVpcTf := serializeVpcPeeringAccepterVPC(ctx, http.AccepterVpc, diags)
	sourceVpcTf := serializeVpcPeeringSourceVPC(ctx, http.SourceVpc, diags)

	var httpExpirationDate, accepterVpcId, sourceVpcId *string
	if http.ExpirationDate != nil {
		tmpDate := *(http.ExpirationDate)
		tmpStr := tmpDate.String()
		httpExpirationDate = &tmpStr
	}
	if http.AccepterVpc != nil {
		tmp := *(http.AccepterVpc)
		accepterVpcId = tmp.VpcId
	}
	if http.SourceVpc != nil {
		tmp := *(http.SourceVpc)
		sourceVpcId = tmp.VpcId
	}

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	return &VpcPeeringModel{
		AccepterVpc:    accepterVpcTf,
		AccepterVpcId:  types.StringPointerValue(accepterVpcId),
		ExpirationDate: types.StringPointerValue(httpExpirationDate),
		Id:             types.StringPointerValue(http.Id),
		SourceVpc:      sourceVpcTf,
		SourceVpcId:    types.StringPointerValue(sourceVpcId),
		State:          vpcPeeringState,
		Tags:           tagsTf,
	}
}

func serializeVpcPeeringAccepterVPC(ctx context.Context, http *numspot.AccepterVpc, diags *diag.Diagnostics) AccepterVpcValue {
	if http == nil {
		return NewAccepterVpcValueNull()
	}

	value, diagnostics := NewAccepterVpcValue(
		AccepterVpcValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"ip_range": types.StringPointerValue(http.IpRange),
			"vpc_id":   types.StringPointerValue(http.VpcId),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func serializeVpcPeeringSourceVPC(ctx context.Context, http *numspot.SourceVpc, diags *diag.Diagnostics) SourceVpcValue {
	if http == nil {
		return NewSourceVpcValueNull()
	}

	value, diagnostics := NewSourceVpcValue(
		SourceVpcValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"ip_range": types.StringPointerValue(http.IpRange),
			"vpc_id":   types.StringPointerValue(http.VpcId),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func serializeVpcPeeringState(ctx context.Context, http *numspot.VpcPeeringState, diags *diag.Diagnostics) StateValue {
	if http == nil {
		return NewStateValueNull()
	}

	value, diagnostics := NewStateValue(
		StateValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			//"message": types.StringPointerValue(http.Message),
			"name": types.StringPointerValue(http.Name),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func deserializeVpcPeering(tf VpcPeeringModel) numspot.CreateVpcPeeringJSONRequestBody {
	return numspot.CreateVpcPeeringJSONRequestBody{
		AccepterVpcId: tf.AccepterVpcId.ValueString(),
		SourceVpcId:   tf.SourceVpcId.ValueString(),
	}
}
