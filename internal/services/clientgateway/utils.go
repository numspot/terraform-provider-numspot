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

func ClientGatewayFromHttpToTf(ctx context.Context, http *numspot.ClientGateway, diags *diag.Diagnostics) *ClientGatewayModel {
	var tagsTf types.List

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	return &ClientGatewayModel{
		BgpAsn:         utils.FromIntPtrToTfInt64(http.BgpAsn),
		ConnectionType: types.StringPointerValue(http.ConnectionType),
		Id:             types.StringPointerValue(http.Id),
		PublicIp:       types.StringPointerValue(http.PublicIp),
		State:          types.StringPointerValue(http.State),
		Tags:           tagsTf,
	}
}

func ClientGatewayFromTfToCreateRequest(tf *ClientGatewayModel) numspot.CreateClientGatewayJSONRequestBody {
	return numspot.CreateClientGatewayJSONRequestBody{
		BgpAsn:         utils.FromTfInt64ToInt(tf.BgpAsn),
		ConnectionType: tf.ConnectionType.ValueString(),
		PublicIp:       tf.PublicIp.ValueString(),
	}
}

func ClientGatewaysFromTfToAPIReadParams(ctx context.Context, tf ClientGatewaysDataSourceModel, diags *diag.Diagnostics) numspot.ReadClientGatewaysParams {
	return numspot.ReadClientGatewaysParams{
		States:          utils.TfStringListToStringPtrList(ctx, tf.States, diags),
		TagKeys:         utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues:       utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
		Tags:            utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
		Ids:             utils.TfStringListToStringPtrList(ctx, tf.IDs, diags),
		ConnectionTypes: utils.TfStringListToStringPtrList(ctx, tf.ConnectionTypes, diags),
		BgpAsns:         utils.TFInt64ListToIntListPointer(ctx, tf.BgpAsns, diags),
		PublicIps:       utils.TfStringListToStringPtrList(ctx, tf.PublicIps, diags),
	}
}

func ClientGatewaysFromHttpToTfDatasource(ctx context.Context, http *numspot.ClientGateway, diags *diag.Diagnostics) *ClientGatewayModel {
	var (
		tagsList types.List
		bgpAsnTf types.Int64
	)

	if http.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
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
	}
}
