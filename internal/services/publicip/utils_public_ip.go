package publicip

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	utils2 "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func PublicIpFromHttpToTf(ctx context.Context, elt *numspot.PublicIp) (*PublicIpModel, diag.Diagnostics) {
	var (
		diags    diag.Diagnostics
		tagsList types.List
	)

	if elt.Tags != nil {
		tagsList, diags = utils2.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *elt.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &PublicIpModel{
		Id:           types.StringPointerValue(elt.Id),
		NicId:        types.StringPointerValue(elt.NicId),
		PrivateIp:    types.StringPointerValue(elt.PrivateIp),
		PublicIp:     types.StringPointerValue(elt.PublicIp),
		VmId:         types.StringPointerValue(elt.VmId),
		LinkPublicIP: types.StringPointerValue(elt.LinkPublicIpId),
		Tags:         tagsList,
	}, nil
}

func invokeLinkPublicIP(ctx context.Context, provider services.IProvider, data *PublicIpModel) (*string, error) {
	var payload numspot.LinkPublicIpJSONRequestBody
	if !data.VmId.IsNull() {
		payload = numspot.LinkPublicIpJSONRequestBody{VmId: data.VmId.ValueStringPointer()}
	} else {
		payload = numspot.LinkPublicIpJSONRequestBody{
			NicId:     data.NicId.ValueStringPointer(),
			PrivateIp: data.PrivateIp.ValueStringPointer(),
		}
	}
	res, err := provider.GetNumspotClient().LinkPublicIpWithResponse(ctx, provider.GetSpaceID(), data.Id.ValueString(), payload)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, utils2.HandleError(res.Body)
	}

	return res.JSON200.LinkPublicIpId, nil
}

func invokeUnlinkPublicIP(ctx context.Context, provider services.IProvider, data *PublicIpModel) error {
	payload := numspot.UnlinkPublicIpJSONRequestBody{
		LinkPublicIpId: data.LinkPublicIP.ValueStringPointer(),
	}
	res, err := provider.GetNumspotClient().UnlinkPublicIpWithResponse(ctx, provider.GetSpaceID(), data.Id.ValueString(), payload)
	if err != nil {
		return err
	}
	if res.StatusCode() != http.StatusOK {
		return utils2.HandleError(res.Body)
	}

	return nil
}

func refreshState(ctx context.Context, provider services.IProvider, id string) (*PublicIpModel, diag.Diagnostics) {
	// Refresh state
	var diags diag.Diagnostics

	res, err := provider.GetNumspotClient().ReadPublicIpsByIdWithResponse(ctx, provider.GetSpaceID(), id)
	if err != nil {
		diags.AddError("Failed to read public ip", err.Error())
		return nil, diags
	}

	if res.StatusCode() != http.StatusOK {
		apiError := utils2.HandleError(res.Body)
		diags.AddError("Failed to read public ip", apiError.Error())
		return nil, diags
	}

	tf, diags := PublicIpFromHttpToTf(ctx, res.JSON200)
	if diags.HasError() {
		return nil, diags
	}

	return tf, diags
}

func PublicIpsFromTfToAPIReadParams(ctx context.Context, tf PublicIpsDataSourceModel) numspot.ReadPublicIpsParams {
	return numspot.ReadPublicIpsParams{
		LinkPublicIpIds: utils2.TfStringListToStringPtrList(ctx, tf.LinkPublicIpIds),
		NicIds:          utils2.TfStringListToStringPtrList(ctx, tf.NicIds),
		PrivateIps:      utils2.TfStringListToStringPtrList(ctx, tf.PrivateIps),
		TagKeys:         utils2.TfStringListToStringPtrList(ctx, tf.TagKeys),
		TagValues:       utils2.TfStringListToStringPtrList(ctx, tf.TagValues),
		Tags:            utils2.TfStringListToStringPtrList(ctx, tf.Tags),
		Ids:             utils2.TfStringListToStringPtrList(ctx, tf.IDs),
		VmIds:           utils2.TfStringListToStringPtrList(ctx, tf.VmIds),
	}
}

func PublicIpsFromHttpToTfDatasource(ctx context.Context, http *numspot.PublicIp) (*PublicIpModelDatasource, diag.Diagnostics) {
	var (
		tagsList types.List
		diag     diag.Diagnostics
	)

	if http.Tags != nil {
		tagsList, diag = utils2.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diag.HasError() {
			return nil, diag
		}
	}

	return &PublicIpModelDatasource{
		Id:             types.StringPointerValue(http.Id),
		NicId:          types.StringPointerValue(http.NicId),
		PrivateIp:      types.StringPointerValue(http.PrivateIp),
		PublicIp:       types.StringPointerValue(http.PublicIp),
		VmId:           types.StringPointerValue(http.VmId),
		LinkPublicIpId: types.StringPointerValue(http.LinkPublicIpId),
		Tags:           tagsList,
	}, nil
}
