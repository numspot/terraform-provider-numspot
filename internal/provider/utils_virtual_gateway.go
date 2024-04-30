package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_virtual_gateway"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_virtual_gateway"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func vpcToVGLinkFromApi(ctx context.Context, from iaas.VpcToVirtualGatewayLink) (resource_virtual_gateway.NetToVirtualGatewayLinksValue, diag.Diagnostics) {
	return resource_virtual_gateway.NewNetToVirtualGatewayLinksValue(
		resource_virtual_gateway.NetToVirtualGatewayLinksValue{}.AttributeTypes(ctx),
		map[string]attr.Value{},
	)
}

func VirtualGatewayFromHttpToTf(ctx context.Context, http *iaas.VirtualGateway) (*resource_virtual_gateway.VirtualGatewayModel, diag.Diagnostics) {
	var (
		diags                      diag.Diagnostics
		tagsTf                     types.List
		netToVirtualGatewaysLinkTd types.List
	)

	if http.VpcToVirtualGatewayLinks != nil {
		netToVirtualGatewaysLinkTd, diags = utils.GenericListToTfListValue(
			ctx,
			resource_virtual_gateway.NetToVirtualGatewayLinksValue{},
			vpcToVGLinkFromApi,
			*http.VpcToVirtualGatewayLinks,
		)
		if diags.HasError() {
			return nil, diags
		}
	} else {
		netToVirtualGatewaysLinkTd = types.ListNull(resource_virtual_gateway.NetToVirtualGatewayLinksValue{}.Type(ctx))
	}

	if http.Tags != nil {
		tagsTf, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &resource_virtual_gateway.VirtualGatewayModel{
		ConnectionType:           types.StringPointerValue(http.ConnectionType),
		Id:                       types.StringPointerValue(http.Id),
		NetToVirtualGatewayLinks: netToVirtualGatewaysLinkTd,
		State:                    types.StringPointerValue(http.State),
		Tags:                     tagsTf,
	}, diags
}

func VirtualGatewayFromTfToCreateRequest(tf resource_virtual_gateway.VirtualGatewayModel) iaas.CreateVirtualGatewayJSONRequestBody {
	return iaas.CreateVirtualGatewayJSONRequestBody{
		ConnectionType: tf.ConnectionType.ValueString(),
	}
}

func VirtualGatewaysFromTfToAPIReadParams(ctx context.Context, tf VirtualGatewaysDataSourceModel) iaas.ReadVirtualGatewaysParams {
	return iaas.ReadVirtualGatewaysParams{
		States:          utils.TfStringListToStringPtrList(ctx, tf.States),
		TagKeys:         utils.TfStringListToStringPtrList(ctx, tf.TagKeys),
		TagValues:       utils.TfStringListToStringPtrList(ctx, tf.TagValues),
		Tags:            utils.TfStringListToStringPtrList(ctx, tf.Tags),
		Ids:             utils.TfStringListToStringPtrList(ctx, tf.IDs),
		ConnectionTypes: utils.TfStringListToStringPtrList(ctx, tf.ConnectionTypes),
		LinkStates:      utils.TfStringListToStringPtrList(ctx, tf.LinkStates),
		LinkVpcIds:      utils.TfStringListToStringPtrList(ctx, tf.LinkVpcIds),
	}
}

func VirtualGatewaysFromHttpToTfDatasource(ctx context.Context, http *iaas.VirtualGateway) (*datasource_virtual_gateway.VirtualGatewayModel, diag.Diagnostics) {
	var (
		netToVirtualGatewayLinks = types.ListNull(datasource_virtual_gateway.NetToVirtualGatewayLinksValue{}.Type(ctx))
		diags                    diag.Diagnostics
		tagsList                 types.List
	)
	if http.VpcToVirtualGatewayLinks != nil {
		netToVirtualGatewayLinks, diags = utils.GenericListToTfListValue(
			ctx,
			datasource_virtual_gateway.NetToVirtualGatewayLinksValue{},
			fromVpcToVirtualGatewayLinkSchemaToTFVpcToVirtualGatewayLinkList,
			*http.VpcToVirtualGatewayLinks,
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

	return &datasource_virtual_gateway.VirtualGatewayModel{
		Id:                       types.StringPointerValue(http.Id),
		State:                    types.StringPointerValue(http.State),
		ConnectionType:           types.StringPointerValue(http.ConnectionType),
		NetToVirtualGatewayLinks: netToVirtualGatewayLinks,
		Tags:                     tagsList,
	}, nil
}

func fromVpcToVirtualGatewayLinkSchemaToTFVpcToVirtualGatewayLinkList(ctx context.Context, http iaas.VpcToVirtualGatewayLink) (datasource_virtual_gateway.NetToVirtualGatewayLinksValue, diag.Diagnostics) {
	return datasource_virtual_gateway.NewNetToVirtualGatewayLinksValue(
		datasource_virtual_gateway.NetToVirtualGatewayLinksValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"state":  types.StringPointerValue(http.State),
			"vpc_id": types.StringPointerValue(http.VpcId),
		})
}
