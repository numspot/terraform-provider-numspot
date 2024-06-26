package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_subnet"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_subnet"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func SubnetFromHttpToTf(ctx context.Context, http *iaas.Subnet) (*resource_subnet.SubnetModel, diag.Diagnostics) {
	var (
		tagsList types.List
		diags    diag.Diagnostics
	)
	if http.Tags != nil {
		tagsList, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &resource_subnet.SubnetModel{
		AvailableIpsCount:    utils.FromIntPtrToTfInt64(http.AvailableIpsCount),
		Id:                   types.StringPointerValue(http.Id),
		IpRange:              types.StringPointerValue(http.IpRange),
		MapPublicIpOnLaunch:  types.BoolPointerValue(http.MapPublicIpOnLaunch),
		VpcId:                types.StringPointerValue(http.VpcId),
		State:                types.StringPointerValue(http.State),
		AvailabilityZoneName: types.StringPointerValue(http.AvailabilityZoneName),
		Tags:                 tagsList,
	}, nil
}

func SubnetFromTfToCreateRequest(tf *resource_subnet.SubnetModel) iaas.CreateSubnetJSONRequestBody {
	return iaas.CreateSubnetJSONRequestBody{
		IpRange:              tf.IpRange.ValueString(),
		VpcId:                tf.VpcId.ValueString(),
		AvailabilityZoneName: tf.AvailabilityZoneName.ValueStringPointer(),
	}
}

func SubnetsFromTfToAPIReadParams(ctx context.Context, tf SubnetsDataSourceModel) iaas.ReadSubnetsParams {
	return iaas.ReadSubnetsParams{
		AvailableIpsCounts:    utils.TFInt64ListToIntListPointer(ctx, tf.AvailableIpsCounts),
		IpRanges:              utils.TfStringListToStringPtrList(ctx, tf.IpRanges),
		States:                utils.TfStringListToStringPtrList(ctx, tf.States),
		VpcIds:                utils.TfStringListToStringPtrList(ctx, tf.VpcIds),
		Ids:                   utils.TfStringListToStringPtrList(ctx, tf.IDs),
		AvailabilityZoneNames: utils.TfStringListToStringPtrList(ctx, tf.AvailabilityZoneNames),
	}
}

func SubnetsFromHttpToTfDatasource(ctx context.Context, http *iaas.Subnet) (*datasource_subnet.SubnetModel, diag.Diagnostics) {
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
