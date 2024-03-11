package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_vpc"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_vpc"
)

func NetFromHttpToTf(ctx context.Context, http *api.Vpc) resource_vpc.VpcModel {
	return resource_vpc.VpcModel{
		DhcpOptionsSetId: types.StringPointerValue(http.DhcpOptionsSetId),
		Id:               types.StringPointerValue(http.Id),
		IpRange:          types.StringPointerValue(http.IpRange),
		State:            types.StringPointerValue(http.State),
		Tenancy:          types.StringPointerValue(http.Tenancy),
		Tags:             types.ListNull(tags.TagsValue{}.Type(ctx)),
	}
}

func NetFromTfToCreateRequest(tf *resource_vpc.VpcModel) api.CreateVpcJSONRequestBody {
	return api.CreateVpcJSONRequestBody{
		IpRange: tf.IpRange.ValueString(),
		Tenancy: tf.Tenancy.ValueStringPointer(),
	}
}

func VPCsFromTfToAPIReadParams(ctx context.Context, tf VPCsDataSourceModel) api.ReadVpcsParams {
	return api.ReadVpcsParams{
		DhcpOptionsSetIds: utils.TfStringListToStringPtrList(ctx, tf.DHCPOptionsSetIds),
		IpRanges:          utils.TfStringListToStringPtrList(ctx, tf.IPRanges),
		IsDefault:         tf.IsDefault.ValueBoolPointer(),
		States:            utils.TfStringListToStringPtrList(ctx, tf.States),
		Ids:               utils.TfStringListToStringPtrList(ctx, tf.IDs),
	}
}

func VPCsFromHttpToTfDatasource(ctx context.Context, http *api.Vpc) (*datasource_vpc.VpcModel, diag.Diagnostics) {
	return &datasource_vpc.VpcModel{
		DhcpOptionsSetId: types.StringPointerValue(http.DhcpOptionsSetId),
		Id:               types.StringPointerValue(http.Id),
		IpRange:          types.StringPointerValue(http.IpRange),
		State:            types.StringPointerValue(http.State),
		Tenancy:          types.StringPointerValue(http.Tenancy),
	}, nil
}
