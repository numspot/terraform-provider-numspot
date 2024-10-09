package internetgateway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func InternetServiceFromTfToHttp(tf InternetGatewayModel) *numspot.InternetGateway {
	return &numspot.InternetGateway{
		Id:    tf.Id.ValueStringPointer(),
		VpcId: tf.VpcId.ValueStringPointer(),
		State: tf.State.ValueStringPointer(),
	}
}

func InternetServiceFromHttpToTf(ctx context.Context, http *numspot.InternetGateway, diags *diag.Diagnostics) *InternetGatewayModel {
	var tagsTf types.List

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	return &InternetGatewayModel{
		Id:    types.StringPointerValue(http.Id),
		VpcId: types.StringPointerValue(http.VpcId),
		State: types.StringPointerValue(http.State),
		Tags:  tagsTf,
	}
}

func InternetGatewaysFromTfToAPIReadParams(ctx context.Context, tf InternetGatewaysDataSourceModel, diags *diag.Diagnostics) numspot.ReadInternetGatewaysParams {
	return numspot.ReadInternetGatewaysParams{
		TagKeys:    utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues:  utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
		Tags:       utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
		Ids:        utils.TfStringListToStringPtrList(ctx, tf.IDs, diags),
		LinkStates: utils.TfStringListToStringPtrList(ctx, tf.LinkStates, diags),
		LinkVpcIds: utils.TfStringListToStringPtrList(ctx, tf.LinkVpcIds, diags),
	}
}

func InternetGatewaysFromHttpToTfDatasource(ctx context.Context, http *numspot.InternetGateway, diags *diag.Diagnostics) *InternetGatewayModel {
	var tagsList types.List

	if http.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return &InternetGatewayModel{
		Id:    types.StringPointerValue(http.Id),
		State: types.StringPointerValue(http.State),
		VpcId: types.StringPointerValue(http.VpcId),
		Tags:  tagsList,
	}
}
