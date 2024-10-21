package subnet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func SubnetFromHttpToTf(ctx context.Context, http *numspot.Subnet, diags *diag.Diagnostics) *SubnetModel {
	var tagsList types.List
	if http.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return &SubnetModel{
		AvailableIpsCount:    utils.FromIntPtrToTfInt64(http.AvailableIpsCount),
		Id:                   types.StringPointerValue(http.Id),
		IpRange:              types.StringPointerValue(http.IpRange),
		MapPublicIpOnLaunch:  types.BoolPointerValue(http.MapPublicIpOnLaunch),
		VpcId:                types.StringPointerValue(http.VpcId),
		State:                types.StringPointerValue(http.State),
		AvailabilityZoneName: types.StringPointerValue(http.AvailabilityZoneName),
		Tags:                 tagsList,
	}
}

func SubnetFromTfToCreateRequest(tf *SubnetModel) numspot.CreateSubnetJSONRequestBody {
	return numspot.CreateSubnetJSONRequestBody{
		IpRange:              tf.IpRange.ValueString(),
		VpcId:                tf.VpcId.ValueString(),
		AvailabilityZoneName: tf.AvailabilityZoneName.ValueStringPointer(),
	}
}

func SubnetsFromTfToAPIReadParams(ctx context.Context, tf SubnetsDataSourceModel, diags *diag.Diagnostics) numspot.ReadSubnetsParams {
	return numspot.ReadSubnetsParams{
		AvailableIpsCounts:    utils.TFInt64ListToIntListPointer(ctx, tf.AvailableIpsCounts, diags),
		IpRanges:              utils.TfStringListToStringPtrList(ctx, tf.IpRanges, diags),
		States:                utils.TfStringListToStringPtrList(ctx, tf.States, diags),
		VpcIds:                utils.TfStringListToStringPtrList(ctx, tf.VpcIds, diags),
		Ids:                   utils.TfStringListToStringPtrList(ctx, tf.Ids, diags),
		AvailabilityZoneNames: utils.TfStringListToStringPtrList(ctx, tf.AvailabilityZoneNames, diags),
	}
}

func SubnetsFromHttpToTfDatasource(ctx context.Context, http *numspot.Subnet, diags *diag.Diagnostics) *SubnetModel {
	var tagsList types.List

	if http.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return &SubnetModel{
		AvailabilityZoneName: types.StringPointerValue(http.AvailabilityZoneName),
		AvailableIpsCount:    utils.FromIntPtrToTfInt64(http.AvailableIpsCount),
		Id:                   types.StringPointerValue(http.Id),
		IpRange:              types.StringPointerValue(http.IpRange),
		MapPublicIpOnLaunch:  types.BoolPointerValue(http.MapPublicIpOnLaunch),
		State:                types.StringPointerValue(http.State),
		VpcId:                types.StringPointerValue(http.VpcId),
		Tags:                 tagsList,
	}
}
