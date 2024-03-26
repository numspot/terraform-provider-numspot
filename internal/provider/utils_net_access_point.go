package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_net_access_point"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func NetAccessPointFromHttpToTf(ctx context.Context, http *iaas.VpcAccessPoint, diagnostics diag.Diagnostics) *resource_net_access_point.NetAccessPointModel {
	routeTablesId, diags := types.ListValueFrom(ctx, types.StringType, http.RouteTableIds)
	if diags.HasError() {
		diagnostics.Append(diags...)
		return nil
	}

	return &resource_net_access_point.NetAccessPointModel{
		Id:            types.StringPointerValue(http.Id),
		NetId:         types.StringPointerValue(http.VpcId),
		RouteTableIds: routeTablesId,
		ServiceName:   types.StringPointerValue(http.ServiceName),
		State:         types.StringPointerValue(http.State),
	}
}

func NetAccessPointFromTfToCreateRequest(ctx context.Context, tf *resource_net_access_point.NetAccessPointModel) iaas.CreateVpcAccessPointJSONRequestBody {
	routeTableIds := utils.TfStringListToStringList(ctx, tf.RouteTableIds)

	return iaas.CreateVpcAccessPointJSONRequestBody{
		VpcId:         tf.NetId.ValueString(),
		RouteTableIds: &routeTableIds,
		ServiceName:   tf.ServiceName.ValueString(),
	}
}
