package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_vpc_peering"
)

func accepterVpcFromApi(ctx context.Context, http *iaas.AccepterVpc) (resource_vpc_peering.AccepterVpcValue, diag.Diagnostics) {
	if http == nil {
		return resource_vpc_peering.NewAccepterVpcValueNull(), nil
	}

	return resource_vpc_peering.NewAccepterVpcValue(
		resource_vpc_peering.AccepterVpcValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"account_id": types.StringPointerValue(http.AccountId),
			"ip_range":   types.StringPointerValue(http.IpRange),
			"vpc_id":     types.StringPointerValue(http.VpcId),
		},
	)
}

func sourceVpcFromApi(ctx context.Context, http *iaas.SourceVpc) (resource_vpc_peering.SourceVpcValue, diag.Diagnostics) {
	if http == nil {
		return resource_vpc_peering.NewSourceVpcValueNull(), nil
	}

	return resource_vpc_peering.NewSourceVpcValue(
		resource_vpc_peering.SourceVpcValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"account_id": types.StringPointerValue(http.AccountId),
			"ip_range":   types.StringPointerValue(http.IpRange),
			"vpc_id":     types.StringPointerValue(http.VpcId),
		},
	)
}

func vpcPeeringStateFromApi(ctx context.Context, http *iaas.VpcPeeringState) (resource_vpc_peering.StateValue, diag.Diagnostics) {
	if http == nil {
		return resource_vpc_peering.NewStateValueNull(), nil
	}

	return resource_vpc_peering.NewStateValue(
		resource_vpc_peering.StateValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"message": types.StringPointerValue(http.Message),
			"name":    types.StringPointerValue(http.Name),
		},
	)
}

func VpcPeeringFromHttpToTf(ctx context.Context, http *iaas.VpcPeering) (*resource_vpc_peering.VpcPeeringModel, diag.Diagnostics) {
	// In the event that the creation of VPC peering fails, the error message might be found in
	// the "state" field. If the state's name is "failed", then the error message will be contained
	// in the state's message. We must address this particular scenario.
	var diagnostics diag.Diagnostics
	vpcPeeringStateHttp := http.State

	if vpcPeeringStateHttp != nil {
		message := vpcPeeringStateHttp.Message
		name := vpcPeeringStateHttp.Name

		if name != nil && *name == "failed" {
			diagnostics.AddError("Failed to create vpc peering", *message)
			return nil, diagnostics
		}
	}

	vpcPeeringState, diagnostics := vpcPeeringStateFromApi(ctx, vpcPeeringStateHttp)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	accepterVpcTf, diagnostics := accepterVpcFromApi(ctx, http.AccepterVpc)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	sourceVpcTf, diagnostics := sourceVpcFromApi(ctx, http.SourceVpc)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

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

	return &resource_vpc_peering.VpcPeeringModel{
		AccepterVpc:    accepterVpcTf,
		AccepterVpcId:  types.StringPointerValue(accepterVpcId),
		ExpirationDate: types.StringPointerValue(httpExpirationDate),
		Id:             types.StringPointerValue(http.Id),
		SourceVpc:      sourceVpcTf,
		SourceVpcId:    types.StringPointerValue(sourceVpcId),
		State:          vpcPeeringState,
	}, diagnostics
}

func VpcPeeringFromTfToCreateRequest(tf resource_vpc_peering.VpcPeeringModel) iaas.CreateVpcPeeringJSONRequestBody {
	return iaas.CreateVpcPeeringJSONRequestBody{
		AccepterVpcId: tf.AccepterVpcId.ValueString(),
		SourceVpcId:   tf.SourceVpcId.ValueString(),
	}
}
