package vpc

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func NetFromHttpToTf(ctx context.Context, http *numspot.Vpc, diags *diag.Diagnostics) *VpcModel {
	var tagsTf types.List

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return &VpcModel{
		DhcpOptionsSetId: types.StringPointerValue(http.DhcpOptionsSetId),
		Id:               types.StringPointerValue(http.Id),
		IpRange:          types.StringPointerValue(http.IpRange),
		State:            types.StringPointerValue(http.State),
		Tenancy:          types.StringPointerValue(http.Tenancy),
		Tags:             tagsTf,
	}
}

func NetFromTfToCreateRequest(tf *VpcModel) numspot.CreateVpcJSONRequestBody {
	return numspot.CreateVpcJSONRequestBody{
		IpRange: tf.IpRange.ValueString(),
		Tenancy: tf.Tenancy.ValueStringPointer(),
	}
}

func VPCsFromTfToAPIReadParams(ctx context.Context, tf VPCsDataSourceModel, diags *diag.Diagnostics) numspot.ReadVpcsParams {
	return numspot.ReadVpcsParams{
		DhcpOptionsSetIds: utils.TfStringListToStringPtrList(ctx, tf.DHCPOptionsSetIds, diags),
		IpRanges:          utils.TfStringListToStringPtrList(ctx, tf.IPRanges, diags),
		IsDefault:         tf.IsDefault.ValueBoolPointer(),
		States:            utils.TfStringListToStringPtrList(ctx, tf.States, diags),
		TagKeys:           utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues:         utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
		Tags:              utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
		Ids:               utils.TfStringListToStringPtrList(ctx, tf.IDs, diags),
	}
}

func VPCsFromHttpToTfDatasource(ctx context.Context, http *numspot.Vpc, diags *diag.Diagnostics) *VpcModel {
	var tagsList types.List

	if http.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return &VpcModel{
		DhcpOptionsSetId: types.StringPointerValue(http.DhcpOptionsSetId),
		Id:               types.StringPointerValue(http.Id),
		IpRange:          types.StringPointerValue(http.IpRange),
		State:            types.StringPointerValue(http.State),
		Tenancy:          types.StringPointerValue(http.Tenancy),
		Tags:             tagsList,
	}
}

func VpcFromTfToUpdaterequest(ctx context.Context, tf *VpcModel) numspot.UpdateVpcJSONRequestBody {
	return numspot.UpdateVpcJSONRequestBody{
		DhcpOptionsSetId: tf.DhcpOptionsSetId.ValueString(),
	}
}
