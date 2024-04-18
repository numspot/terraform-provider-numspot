package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_vpc"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_vpc"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func NetFromHttpToTf(ctx context.Context, http *iaas.Vpc) (*resource_vpc.VpcModel, diag.Diagnostics) {
	/*var (
		tagsTf types.List
		diags  diag.Diagnostics
	)
	if http.Tags != nil {
		tagsTf, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}*/

	return &resource_vpc.VpcModel{
		DhcpOptionsSetId: types.StringPointerValue(http.DhcpOptionsSetId),
		Id:               types.StringPointerValue(http.Id),
		IpRange:          types.StringPointerValue(http.IpRange),
		State:            types.StringPointerValue(http.State),
		Tenancy:          types.StringPointerValue(http.Tenancy),
		//Tags:             tagsTf,
	}, nil
}

func NetFromTfToCreateRequest(tf *resource_vpc.VpcModel) iaas.CreateVpcJSONRequestBody {
	return iaas.CreateVpcJSONRequestBody{
		IpRange: tf.IpRange.ValueString(),
		Tenancy: tf.Tenancy.ValueStringPointer(),
	}
}

func VPCsFromTfToAPIReadParams(ctx context.Context, tf VPCsDataSourceModel) iaas.ReadVpcsParams {
	return iaas.ReadVpcsParams{
		DhcpOptionsSetIds: utils.TfStringListToStringPtrList(ctx, tf.DHCPOptionsSetIds),
		IpRanges:          utils.TfStringListToStringPtrList(ctx, tf.IPRanges),
		IsDefault:         tf.IsDefault.ValueBoolPointer(),
		States:            utils.TfStringListToStringPtrList(ctx, tf.States),
		//TagKeys:           utils.TfStringListToStringPtrList(ctx, tf.TagKeys),
		//TagValues:         utils.TfStringListToStringPtrList(ctx, tf.TagValues),
		//Tags:              utils.TfStringListToStringPtrList(ctx, tf.Tags),
		Ids: utils.TfStringListToStringPtrList(ctx, tf.IDs),
	}
}

func VPCsFromHttpToTfDatasource(ctx context.Context, http *iaas.Vpc) (*datasource_vpc.VpcModel, diag.Diagnostics) {
	/*var (
		tagsList types.List
		diags    diag.Diagnostics
	)
	if http.Tags != nil {
		tagsList, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}*/
	return &datasource_vpc.VpcModel{
		DhcpOptionsSetId: types.StringPointerValue(http.DhcpOptionsSetId),
		Id:               types.StringPointerValue(http.Id),
		IpRange:          types.StringPointerValue(http.IpRange),
		State:            types.StringPointerValue(http.State),
		Tenancy:          types.StringPointerValue(http.Tenancy),
		//Tags:             tagsList,
	}, nil
}
