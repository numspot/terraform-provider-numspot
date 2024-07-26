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

func accepterVpcFromApi(ctx context.Context, http *numspot.AccepterVpc) (AccepterVpcValue, diag.Diagnostics) {
	if http == nil {
		return NewAccepterVpcValueNull(), nil
	}

	return NewAccepterVpcValue(
		AccepterVpcValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"ip_range": types.StringPointerValue(http.IpRange),
			"vpc_id":   types.StringPointerValue(http.VpcId),
		},
	)
}

func sourceVpcFromApi(ctx context.Context, http *numspot.SourceVpc) (SourceVpcValue, diag.Diagnostics) {
	if http == nil {
		return NewSourceVpcValueNull(), nil
	}

	return NewSourceVpcValue(
		SourceVpcValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"ip_range": types.StringPointerValue(http.IpRange),
			"vpc_id":   types.StringPointerValue(http.VpcId),
		},
	)
}

func vpcPeeringStateFromApi(ctx context.Context, http *numspot.VpcPeeringState) (StateValue, diag.Diagnostics) {
	if http == nil {
		return NewStateValueNull(), nil
	}

	return NewStateValue(
		StateValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"message": types.StringPointerValue(http.Message),
			"name":    types.StringPointerValue(http.Name),
		},
	)
}

func VpcPeeringFromHttpToTf(ctx context.Context, http *numspot.VpcPeering) (*VpcPeeringModel, diag.Diagnostics) {
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
			var errorMessage string
			if message != nil {
				errorMessage = *message
			}
			diags.AddError("Failed to create vpc peering", errorMessage)
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

	return &VpcPeeringModel{
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

func VpcPeeringFromTfToCreateRequest(tf VpcPeeringModel) numspot.CreateVpcPeeringJSONRequestBody {
	return numspot.CreateVpcPeeringJSONRequestBody{
		AccepterVpcId: tf.AccepterVpcId.ValueString(),
		SourceVpcId:   tf.SourceVpcId.ValueString(),
	}
}

func VpcPeeringsFromTfToAPIReadParams(ctx context.Context, tf VpcPeeringsDataSourceModel) numspot.ReadVpcPeeringsParams {
	expirationDates := utils.TfStringListToTimeList(ctx, tf.ExpirationDates, "2020-06-30T00:00:00.000Z")

	return numspot.ReadVpcPeeringsParams{
		ExpirationDates:     &expirationDates,
		StateMessages:       utils.TfStringListToStringPtrList(ctx, tf.StateMessages),
		StateNames:          utils.TfStringListToStringPtrList(ctx, tf.StateNames),
		AccepterVpcIpRanges: utils.TfStringListToStringPtrList(ctx, tf.AccepterVpcIpRanges),
		AccepterVpcVpcIds:   utils.TfStringListToStringPtrList(ctx, tf.AccepterVpcVpcIds),
		Ids:                 utils.TfStringListToStringPtrList(ctx, tf.Ids),
		SourceVpcIpRanges:   utils.TfStringListToStringPtrList(ctx, tf.SourceVpcIpRanges),
		SourceVpcVpcIds:     utils.TfStringListToStringPtrList(ctx, tf.SourceVpcVpcIds),
		TagKeys:             utils.TfStringListToStringPtrList(ctx, tf.TagKeys),
		TagValues:           utils.TfStringListToStringPtrList(ctx, tf.TagValues),
		Tags:                utils.TfStringListToStringPtrList(ctx, tf.Tags),
	}
}

func VpcPeeringsFromHttpToTfDatasource(ctx context.Context, http *numspot.VpcPeering) (*VpcPeeringDatasourceItemModel, diag.Diagnostics) {
	var (
		diags            diag.Diagnostics
		tagsList         types.List
		accepterVpc      AccepterVpcValue
		sourceVpc        SourceVpcValue
		state            StateValue
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
		accepterVpc, diags = NewAccepterVpcValue(AccepterVpcValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"ip_range": types.StringPointerValue(http.AccepterVpc.IpRange),
				"vpc_id":   types.StringPointerValue(http.AccepterVpc.VpcId),
			})
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.SourceVpc != nil {
		sourceVpc, diags = NewSourceVpcValue(SourceVpcValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"ip_range": types.StringPointerValue(http.SourceVpc.IpRange),
				"vpc_id":   types.StringPointerValue(http.SourceVpc.VpcId),
			})
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.State != nil {
		state, diags = NewStateValue(StateValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"message": types.StringPointerValue(http.State.Message),
				"name":    types.StringPointerValue(http.State.Name),
			})
		if diags.HasError() {
			return nil, diags
		}
	}

	return &VpcPeeringDatasourceItemModel{
		Id:             types.StringPointerValue(http.Id),
		Tags:           tagsList,
		AccepterVpc:    accepterVpc,
		ExpirationDate: expirationDateTf,
		SourceVpc:      sourceVpc,
		State:          state,
	}, nil
}
