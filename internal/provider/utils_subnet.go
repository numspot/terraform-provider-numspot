package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_subnet"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_subnet"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func SubnetFromHttpToTf(http *api.Subnet) resource_subnet.SubnetModel {
	return resource_subnet.SubnetModel{
		AvailableIpsCount:    utils.FromIntPtrToTfInt64(http.AvailableIpsCount),
		Id:                   types.StringPointerValue(http.Id),
		IpRange:              types.StringPointerValue(http.IpRange),
		MapPublicIpOnLaunch:  types.BoolPointerValue(http.MapPublicIpOnLaunch),
		VpcId:                types.StringPointerValue(http.VpcId),
		State:                types.StringPointerValue(http.State),
		AvailabilityZoneName: types.StringPointerValue(http.AvailabilityZoneName),
	}
}

func SubnetFromTfToCreateRequest(tf *resource_subnet.SubnetModel) api.CreateSubnetJSONRequestBody {
	return api.CreateSubnetJSONRequestBody{
		IpRange:              tf.IpRange.ValueString(),
		VpcId:                tf.VpcId.ValueString(),
		AvailabilityZoneName: tf.AvailabilityZoneName.ValueStringPointer(),
	}
}

func SubnetsFromTfToAPIReadParams(ctx context.Context, tf SubnetsDataSourceModel) api.ReadSubnetsParams {
	return api.ReadSubnetsParams{
		AvailableIpsCounts:    utils.TFInt64ListToIntListPointer(ctx, tf.AvailableIpsCounts),
		IpRanges:              utils.TfStringListToStringPtrList(ctx, tf.IpRanges),
		States:                utils.TfStringListToStringPtrList(ctx, tf.States),
		VpcIds:                utils.TfStringListToStringPtrList(ctx, tf.VpcIds),
		Ids:                   utils.TfStringListToStringPtrList(ctx, tf.IDs),
		AvailabilityZoneNames: utils.TfStringListToStringPtrList(ctx, tf.AvailabilityZoneNames),
	}
}

func SubnetsFromHttpToTfDatasource(ctx context.Context, http *api.Subnet) (*datasource_subnet.SubnetModel, diag.Diagnostics) {
	return &datasource_subnet.SubnetModel{
		AvailabilityZoneName: types.StringPointerValue(http.AvailabilityZoneName),
		AvailableIpsCount:    utils.FromIntPtrToTfInt64(http.AvailableIpsCount),
		Id:                   types.StringPointerValue(http.Id),
		IpRange:              types.StringPointerValue(http.IpRange),
		MapPublicIpOnLaunch:  types.BoolPointerValue(http.MapPublicIpOnLaunch),
		State:                types.StringPointerValue(http.State),
		VpcId:                types.StringPointerValue(http.VpcId),
	}, nil
}
