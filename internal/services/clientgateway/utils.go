package clientgateway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func ClientGatewayFromTfToHttp(tf *ClientGatewayModel) *numspot.ClientGateway {
	return &numspot.ClientGateway{
		BgpAsn:         utils.FromTfInt64ToIntPtr(tf.BgpAsn),
		ConnectionType: tf.ConnectionType.ValueStringPointer(),
		Id:             tf.Id.ValueStringPointer(),
		PublicIp:       tf.PublicIp.ValueStringPointer(),
		State:          tf.State.ValueStringPointer(),
	}
}

func ClientGatewayFromHttpToTf(ctx context.Context, http *numspot.ClientGateway) (*ClientGatewayModel, diag.Diagnostics) {
	var (
		tagsTf types.List
		diags  diag.Diagnostics
	)

	if http.Tags != nil {
		tagsTf, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &ClientGatewayModel{
		BgpAsn:         utils.FromIntPtrToTfInt64(http.BgpAsn),
		ConnectionType: types.StringPointerValue(http.ConnectionType),
		Id:             types.StringPointerValue(http.Id),
		PublicIp:       types.StringPointerValue(http.PublicIp),
		State:          types.StringPointerValue(http.State),
		Tags:           tagsTf,
	}, diags
}

func ClientGatewayFromTfToCreateRequest(tf *ClientGatewayModel) numspot.CreateClientGatewayJSONRequestBody {
	return numspot.CreateClientGatewayJSONRequestBody{
		BgpAsn:         utils.FromTfInt64ToInt(tf.BgpAsn),
		ConnectionType: tf.ConnectionType.ValueString(),
		PublicIp:       tf.PublicIp.ValueString(),
	}
}

func ClientGatewaysFromTfToAPIReadParams(ctx context.Context, tf ClientGatewaysDataSourceModel) numspot.ReadClientGatewaysParams {
	return numspot.ReadClientGatewaysParams{
		States:          utils.TfStringListToStringPtrList(ctx, tf.States),
		TagKeys:         utils.TfStringListToStringPtrList(ctx, tf.TagKeys),
		TagValues:       utils.TfStringListToStringPtrList(ctx, tf.TagValues),
		Tags:            utils.TfStringListToStringPtrList(ctx, tf.Tags),
		Ids:             utils.TfStringListToStringPtrList(ctx, tf.IDs),
		ConnectionTypes: utils.TfStringListToStringPtrList(ctx, tf.ConnectionTypes),
		BgpAsns:         utils.TFInt64ListToIntListPointer(ctx, tf.BgpAsns),
		PublicIps:       utils.TfStringListToStringPtrList(ctx, tf.PublicIps),
	}
}

func ClientGatewaysFromHttpToTfDatasource(ctx context.Context, http *numspot.ClientGateway) (*ClientGatewayModel, diag.Diagnostics) {
	var (
		diags    diag.Diagnostics
		tagsList types.List
		bgpAsnTf types.Int64
	)

	if http.Tags != nil {
		tagsList, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.BgpAsn != nil {
		bgpAsn := int64(*http.BgpAsn)
		bgpAsnTf = types.Int64PointerValue(&bgpAsn)
	}

	return &ClientGatewayModel{
		Id:             types.StringPointerValue(http.Id),
		State:          types.StringPointerValue(http.State),
		ConnectionType: types.StringPointerValue(http.ConnectionType),
		Tags:           tagsList,
		BgpAsn:         bgpAsnTf,
		PublicIp:       types.StringPointerValue(http.PublicIp),
	}, nil
}
