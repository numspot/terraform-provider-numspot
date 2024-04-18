package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_nat_gateway"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_nat_gateway"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func publicIpFromApi(ctx context.Context, elt iaas.PublicIpLight) (resource_nat_gateway.PublicIpsValue, diag.Diagnostics) {
	return resource_nat_gateway.NewPublicIpsValue(
		resource_nat_gateway.PublicIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"public_ip":    types.StringPointerValue(elt.PublicIp),
			"public_ip_id": types.StringPointerValue(elt.PublicIpId),
		},
	)
}

func NatGatewayFromHttpToTf(ctx context.Context, http *iaas.NatGateway) (*resource_nat_gateway.NatGatewayModel, diag.Diagnostics) {
	var publicIp []iaas.PublicIpLight
	if http.PublicIps != nil {
		publicIp = *http.PublicIps
	}
	// Public Ips
	publicIpsTf, diagnostics := utils.GenericListToTfListValue(
		ctx,
		resource_nat_gateway.PublicIpsValue{},
		publicIpFromApi,
		publicIp,
	)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	// PublicIpId must be the id of the first public io
	var publicIpId *string
	if len(publicIp) > 0 {
		publicIpId = publicIp[0].PublicIpId
	} else {
		publicIpId = nil
	}

	return &resource_nat_gateway.NatGatewayModel{
		Id:         types.StringPointerValue(http.Id),
		PublicIpId: types.StringPointerValue(publicIpId),
		PublicIps:  publicIpsTf,
		State:      types.StringPointerValue(http.State),
		SubnetId:   types.StringPointerValue(http.SubnetId),
		VpcId:      types.StringPointerValue(http.VpcId),
	}, nil
}

func NatGatewayFromTfToCreateRequest(tf resource_nat_gateway.NatGatewayModel) iaas.CreateNatGatewayJSONRequestBody {
	return iaas.CreateNatGatewayJSONRequestBody{
		PublicIpId: tf.PublicIpId.ValueString(),
		SubnetId:   tf.SubnetId.ValueString(),
	}
}

func NatGatewaysFromTfToAPIReadParams(ctx context.Context, tf NatGatewaysDataSourceModel) iaas.ReadNatGatewayParams {
	return iaas.ReadNatGatewayParams{
		SubnetIds: utils.TfStringListToStringPtrList(ctx, tf.SubnetIds),
		VpcIds:    utils.TfStringListToStringPtrList(ctx, tf.VpcIds),
		States:    utils.TfStringListToStringPtrList(ctx, tf.States),
		TagKeys:   utils.TfStringListToStringPtrList(ctx, tf.TagKeys),
		TagValues: utils.TfStringListToStringPtrList(ctx, tf.TagValues),
		Tags:      utils.TfStringListToStringPtrList(ctx, tf.Tags),
		Ids:       utils.TfStringListToStringPtrList(ctx, tf.IDs),
	}
}

func fromPublicIpSchemaToTFPublicIpsList(ctx context.Context, http iaas.PublicIpLight) (resource_nat_gateway.PublicIpsValue, diag.Diagnostics) {
	return resource_nat_gateway.NewPublicIpsValue(
		resource_nat_gateway.PublicIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"public_ip":    types.StringPointerValue(http.PublicIp),
			"public_ip_id": types.StringPointerValue(http.PublicIpId),
		})
}

func NatGatewaysFromHttpToTfDatasource(ctx context.Context, http *iaas.NatGateway) (*datasource_nat_gateway.NatGatewayModel, diag.Diagnostics) {
	var (
		publicIps = types.ListNull(resource_nat_gateway.PublicIpsValue{}.Type(ctx))
		diags     diag.Diagnostics
		tagsList  types.List
	)
	if http.PublicIps != nil {
		publicIps, diags = utils.GenericListToTfListValue(
			ctx,
			resource_nat_gateway.PublicIpsValue{},
			fromPublicIpSchemaToTFPublicIpsList,
			*http.PublicIps,
		)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.Tags != nil {
		tagsList, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &datasource_nat_gateway.NatGatewayModel{
		Id:        types.StringPointerValue(http.Id),
		State:     types.StringPointerValue(http.State),
		PublicIps: publicIps,
		SubnetId:  types.StringPointerValue(http.SubnetId),
		VpcId:     types.StringPointerValue(http.VpcId),
		Tags:      tagsList,
	}, nil
}
