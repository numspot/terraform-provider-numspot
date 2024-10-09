package vpcpeering

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func accepterVpcFromApi(ctx context.Context, http *numspot.AccepterVpc, diags *diag.Diagnostics) AccepterVpcValue {
	if http == nil {
		return NewAccepterVpcValueNull()
	}

	value, diagnostics := NewAccepterVpcValue(
		AccepterVpcValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"ip_range": types.StringPointerValue(http.IpRange),
			"vpc_id":   types.StringPointerValue(http.VpcId),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func sourceVpcFromApi(ctx context.Context, http *numspot.SourceVpc, diags *diag.Diagnostics) SourceVpcValue {
	if http == nil {
		return NewSourceVpcValueNull()
	}

	value, diagnostics := NewSourceVpcValue(
		SourceVpcValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"ip_range": types.StringPointerValue(http.IpRange),
			"vpc_id":   types.StringPointerValue(http.VpcId),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func vpcPeeringStateFromApi(ctx context.Context, http *numspot.VpcPeeringState, diags *diag.Diagnostics) StateValue {
	if http == nil {
		return NewStateValueNull()
	}

	value, diagnostics := NewStateValue(
		StateValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"message": types.StringPointerValue(http.Message),
			"name":    types.StringPointerValue(http.Name),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func VpcPeeringFromHttpToTf(ctx context.Context, http *numspot.VpcPeering, diags *diag.Diagnostics) *VpcPeeringModel {
	// In the event that the creation of VPC peering fails, the error message might be found in
	// the "state" field. If the state's name is "failed", then the error message will be contained
	// in the state's message. We must address this particular scenario.
	var tagsTf types.List

	vpcPeeringStateHttp := http.State

	if vpcPeeringStateHttp != nil {
		message := vpcPeeringStateHttp.Message
		name := vpcPeeringStateHttp.Name

		if name != nil && *name == "failed" {
			var errorMessage string
			if message != nil {
				errorMessage = *message
			}
			diags.AddError("Failed to create vpc peering", errorMessage)
			return nil
		}
	}

	vpcPeeringState := vpcPeeringStateFromApi(ctx, vpcPeeringStateHttp, diags)
	accepterVpcTf := accepterVpcFromApi(ctx, http.AccepterVpc, diags)
	sourceVpcTf := sourceVpcFromApi(ctx, http.SourceVpc, diags)

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
		tagsTf = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	return &VpcPeeringModel{
		AccepterVpc:    accepterVpcTf,
		AccepterVpcId:  types.StringPointerValue(accepterVpcId),
		ExpirationDate: types.StringPointerValue(httpExpirationDate),
		Id:             types.StringPointerValue(http.Id),
		SourceVpc:      sourceVpcTf,
		SourceVpcId:    types.StringPointerValue(sourceVpcId),
		State:          vpcPeeringState,
		Tags:           tagsTf,
	}
}

func VpcPeeringFromTfToCreateRequest(tf VpcPeeringModel) numspot.CreateVpcPeeringJSONRequestBody {
	return numspot.CreateVpcPeeringJSONRequestBody{
		AccepterVpcId: tf.AccepterVpcId.ValueString(),
		SourceVpcId:   tf.SourceVpcId.ValueString(),
	}
}

func VpcPeeringsFromTfToAPIReadParams(ctx context.Context, tf VpcPeeringsDataSourceModel, diags *diag.Diagnostics) numspot.ReadVpcPeeringsParams {
	expirationDates := utils.TfStringListToTimeList(ctx, tf.ExpirationDates, "2020-06-30T00:00:00.000Z", diags)

	return numspot.ReadVpcPeeringsParams{
		ExpirationDates:     &expirationDates,
		StateMessages:       utils.TfStringListToStringPtrList(ctx, tf.StateMessages, diags),
		StateNames:          utils.TfStringListToStringPtrList(ctx, tf.StateNames, diags),
		AccepterVpcIpRanges: utils.TfStringListToStringPtrList(ctx, tf.AccepterVpcIpRanges, diags),
		AccepterVpcVpcIds:   utils.TfStringListToStringPtrList(ctx, tf.AccepterVpcVpcIds, diags),
		Ids:                 utils.TfStringListToStringPtrList(ctx, tf.Ids, diags),
		SourceVpcIpRanges:   utils.TfStringListToStringPtrList(ctx, tf.SourceVpcIpRanges, diags),
		SourceVpcVpcIds:     utils.TfStringListToStringPtrList(ctx, tf.SourceVpcVpcIds, diags),
		TagKeys:             utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues:           utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
		Tags:                utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
	}
}

func VpcPeeringsFromHttpToTfDatasource(ctx context.Context, http *numspot.VpcPeering, diags *diag.Diagnostics) *VpcPeeringDatasourceItemModel {
	var (
		tagsList         types.List
		accepterVpc      AccepterVpcValue
		sourceVpc        SourceVpcValue
		state            StateValue
		expirationDateTf types.String
	)

	if http.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	if http.ExpirationDate != nil {
		expirationDate := *http.ExpirationDate
		expirationDateTf = types.StringValue(expirationDate.Format(time.RFC3339))
	}

	if http.AccepterVpc != nil {
		var diagnostics diag.Diagnostics
		accepterVpc, diagnostics = NewAccepterVpcValue(AccepterVpcValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"ip_range": types.StringPointerValue(http.AccepterVpc.IpRange),
				"vpc_id":   types.StringPointerValue(http.AccepterVpc.VpcId),
			})
		diags.Append(diagnostics...)
	}

	if http.SourceVpc != nil {
		var diagnostics diag.Diagnostics
		sourceVpc, diagnostics = NewSourceVpcValue(SourceVpcValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"ip_range": types.StringPointerValue(http.SourceVpc.IpRange),
				"vpc_id":   types.StringPointerValue(http.SourceVpc.VpcId),
			})
		diags.Append(diagnostics...)
	}

	if http.State != nil {
		var diagnostics diag.Diagnostics
		state, diagnostics = NewStateValue(StateValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"message": types.StringPointerValue(http.State.Message),
				"name":    types.StringPointerValue(http.State.Name),
			})
		diags.Append(diagnostics...)
	}

	return &VpcPeeringDatasourceItemModel{
		Id:             types.StringPointerValue(http.Id),
		Tags:           tagsList,
		AccepterVpc:    accepterVpc,
		ExpirationDate: expirationDateTf,
		SourceVpc:      sourceVpc,
		State:          state,
	}
}
