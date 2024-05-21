package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_vpc_peering"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_vpc_peering"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func accepterVpcFromApi(ctx context.Context, http *iaas.AccepterVpc) (resource_vpc_peering.AccepterVpcValue, diag.Diagnostics) {
	if http == nil {
		return resource_vpc_peering.NewAccepterVpcValueNull(), nil
	}

	return resource_vpc_peering.NewAccepterVpcValue(
		resource_vpc_peering.AccepterVpcValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"ip_range": types.StringPointerValue(http.IpRange),
			"vpc_id":   types.StringPointerValue(http.VpcId),
		},
	)
}

func sourceVpcFromApi(ctx context.Context, http *iaas.SourceVpc) (resource_vpc_peering.SourceVpcValue, diag.Diagnostics) {
	if http == nil {
		return resource_vpc_peering.NewSourceVpcValueNull(), nil
	}

	return resource_vpc_peering.NewSourceVpcValue(
		resource_vpc_peering.SourceVpcValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"ip_range": types.StringPointerValue(http.IpRange),
			"vpc_id":   types.StringPointerValue(http.VpcId),
		},
	)
}

func vpcPeeringStateFromApi(ctx context.Context, http *iaas.VpcPeeringState) (resource_vpc_peering.StateValue, diag.Diagnostics) {
	if http == nil {
		return resource_vpc_peering.NewStateValueNull(), nil
	}

	return resource_vpc_peering.NewStateValue(
		resource_vpc_peering.StateValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"message": types.StringPointerValue(http.Message),
			"name":    types.StringPointerValue(http.Name),
		},
	)
}

func VpcPeeringFromHttpToTf(ctx context.Context, http *iaas.VpcPeering) (*resource_vpc_peering.VpcPeeringModel, diag.Diagnostics) {
	// In the event that the creation of VPC peering fails, the error message might be found in
	// the "state" field. If the state's name is "failed", then the error message will be contained
	// in the state's message. We must address this particular scenario.
	var (
		tagsTf types.List
		diags  diag.Diagnostics
	)
	vpcPeeringStateHttp := http.State

	if vpcPeeringStateHttp != nil {
		message := vpcPeeringStateHttp.Message
		name := vpcPeeringStateHttp.Name

		if name != nil && *name == "failed" {
			diags.AddError("Failed to create vpc peering", *message)
			return nil, diags
		}
	}

	vpcPeeringState, diags := vpcPeeringStateFromApi(ctx, vpcPeeringStateHttp)
	if diags.HasError() {
		return nil, diags
	}

	accepterVpcTf, diags := accepterVpcFromApi(ctx, http.AccepterVpc)
	if diags.HasError() {
		return nil, diags
	}

	sourceVpcTf, diags := sourceVpcFromApi(ctx, http.SourceVpc)
	if diags.HasError() {
		return nil, diags
	}

	var httpExpirationDate, accepterVpcId, sourceVpcId *string
	if http.ExpirationDate != nil {
		tmpDate := *(http.ExpirationDate)
		tmpStr := tmpDate.String()
		httpExpirationDate = &tmpStr
	}
	if http.AccepterVpc != nil {
		tmp := *(http.AccepterVpc)
		accepterVpcId = tmp.VpcId
	}
	if http.SourceVpc != nil {
		tmp := *(http.SourceVpc)
		sourceVpcId = tmp.VpcId
	}

	if http.Tags != nil {
		tagsTf, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &resource_vpc_peering.VpcPeeringModel{
		AccepterVpc:    accepterVpcTf,
		AccepterVpcId:  types.StringPointerValue(accepterVpcId),
		ExpirationDate: types.StringPointerValue(httpExpirationDate),
		Id:             types.StringPointerValue(http.Id),
		SourceVpc:      sourceVpcTf,
		SourceVpcId:    types.StringPointerValue(sourceVpcId),
		State:          vpcPeeringState,
		Tags:           tagsTf,
	}, diags
}

func VpcPeeringFromTfToCreateRequest(tf resource_vpc_peering.VpcPeeringModel) iaas.CreateVpcPeeringJSONRequestBody {
	return iaas.CreateVpcPeeringJSONRequestBody{
		AccepterVpcId: tf.AccepterVpcId.ValueString(),
		SourceVpcId:   tf.SourceVpcId.ValueString(),
	}
}

func VpcPeeringsFromTfToAPIReadParams(ctx context.Context, tf VpcPeeringsDataSourceModel) iaas.ReadVpcPeeringsParams {
	expirationDates := utils.TfStringListToTimeList(ctx, tf.ExpirationDates, "2020-06-30T00:00:00.000Z")

	return iaas.ReadVpcPeeringsParams{
		ExpirationDates:       &expirationDates,
		StateMessages:         utils.TfStringListToStringPtrList(ctx, tf.StateMessages),
		StateNames:            utils.TfStringListToStringPtrList(ctx, tf.StateNames),
		AccepterVpcAccountIds: utils.TfStringListToStringPtrList(ctx, tf.AccepterVpcAccountIds),
		AccepterVpcIpRanges:   utils.TfStringListToStringPtrList(ctx, tf.AccepterVpcIpRanges),
		AccepterVpcVpcIds:     utils.TfStringListToStringPtrList(ctx, tf.AccepterVpcVpcIds),
		Ids:                   utils.TfStringListToStringPtrList(ctx, tf.IDs),
		SourceVpcAccountIds:   utils.TfStringListToStringPtrList(ctx, tf.SourceVpcAccountIds),
		SourceVpcIpRanges:     utils.TfStringListToStringPtrList(ctx, tf.SourceVpcIpRanges),
		SourceVpcVpcIds:       utils.TfStringListToStringPtrList(ctx, tf.SourceVpcVpcIds),
		TagKeys:               utils.TfStringListToStringPtrList(ctx, tf.TagKeys),
		TagValues:             utils.TfStringListToStringPtrList(ctx, tf.TagValues),
		Tags:                  utils.TfStringListToStringPtrList(ctx, tf.Tags),
	}
}

func VpcPeeringsFromHttpToTfDatasource(ctx context.Context, http *iaas.VpcPeering) (*datasource_vpc_peering.VpcPeeringModel, diag.Diagnostics) {
	var (
		diags            diag.Diagnostics
		tagsList         types.List
		accepterVpc      datasource_vpc_peering.AccepterVpcValue
		sourceVpc        datasource_vpc_peering.SourceVpcValue
		state            datasource_vpc_peering.StateValue
		expirationDateTf types.String
	)

	if http.Tags != nil {
		tagsList, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.ExpirationDate != nil {
		expirationDate := *http.ExpirationDate
		expirationDateTf = types.StringValue(expirationDate.Format(time.RFC3339))
	}

	if http.AccepterVpc != nil {
		accepterVpc, diags = datasource_vpc_peering.NewAccepterVpcValue(datasource_vpc_peering.AccepterVpcValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"ip_range": types.StringPointerValue(http.AccepterVpc.IpRange),
				"vpc_id":   types.StringPointerValue(http.AccepterVpc.VpcId),
			})
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.SourceVpc != nil {
		sourceVpc, diags = datasource_vpc_peering.NewSourceVpcValue(datasource_vpc_peering.SourceVpcValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"ip_range": types.StringPointerValue(http.SourceVpc.IpRange),
				"vpc_id":   types.StringPointerValue(http.SourceVpc.VpcId),
			})
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.State != nil {
		state, diags = datasource_vpc_peering.NewStateValue(datasource_vpc_peering.StateValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"message": types.StringPointerValue(http.State.Message),
				"name":    types.StringPointerValue(http.State.Name),
			})
		if diags.HasError() {
			return nil, diags
		}
	}

	return &datasource_vpc_peering.VpcPeeringModel{
		Id:             types.StringPointerValue(http.Id),
		Tags:           tagsList,
		AccepterVpc:    accepterVpc,
		ExpirationDate: expirationDateTf,
		SourceVpc:      sourceVpc,
		State:          state,
	}, nil
}
