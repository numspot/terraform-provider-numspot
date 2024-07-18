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

func InternetServiceFromHttpToTf(ctx context.Context, http *numspot.InternetGateway) (*InternetGatewayModel, diag.Diagnostics) {
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

	return &InternetGatewayModel{
		Id:    types.StringPointerValue(http.Id),
		VpcId: types.StringPointerValue(http.VpcId),
		State: types.StringPointerValue(http.State),
		Tags:  tagsTf,
	}, diags
}

func InternetGatewaysFromTfToAPIReadParams(ctx context.Context, tf InternetGatewaysDataSourceModel) numspot.ReadInternetGatewaysParams {
	return numspot.ReadInternetGatewaysParams{
		TagKeys:    utils.TfStringListToStringPtrList(ctx, tf.TagKeys),
		TagValues:  utils.TfStringListToStringPtrList(ctx, tf.TagValues),
		Tags:       utils.TfStringListToStringPtrList(ctx, tf.Tags),
		Ids:        utils.TfStringListToStringPtrList(ctx, tf.IDs),
		LinkStates: utils.TfStringListToStringPtrList(ctx, tf.LinkStates),
		LinkVpcIds: utils.TfStringListToStringPtrList(ctx, tf.LinkVpcIds),
	}
}

func InternetGatewaysFromHttpToTfDatasource(ctx context.Context, http *numspot.InternetGateway) (*InternetGatewayModel, diag.Diagnostics) {
	var (
		diags    diag.Diagnostics
		tagsList types.List
	)

	if http.Tags != nil {
		tagsList, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &InternetGatewayModel{
		Id:    types.StringPointerValue(http.Id),
		State: types.StringPointerValue(http.State),
		VpcId: types.StringPointerValue(http.VpcId),
		Tags:  tagsList,
	}, nil
}
