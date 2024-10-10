package publicip

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func PublicIpFromHttpToTf(ctx context.Context, elt *numspot.PublicIp, diags *diag.Diagnostics) *PublicIpModel {
	var tagsList types.List

	if elt.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *elt.Tags, diags)
		if diags.HasError() {
			return nil
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
	}
}

func invokeLinkPublicIP(ctx context.Context, provider *client.NumSpotSDK, data *PublicIpModel) (*string, error) {
	var payload numspot.LinkPublicIpJSONRequestBody
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	if !data.VmId.IsNull() {
		payload = numspot.LinkPublicIpJSONRequestBody{VmId: data.VmId.ValueStringPointer()}
	} else {
		payload = numspot.LinkPublicIpJSONRequestBody{
			NicId:     data.NicId.ValueStringPointer(),
			PrivateIp: data.PrivateIp.ValueStringPointer(),
		}
	}
	res, err := numspotClient.LinkPublicIpWithResponse(ctx, provider.SpaceID, data.Id.ValueString(), payload)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, utils.HandleError(res.Body)
	}

	return res.JSON200.LinkPublicIpId, nil
}

func invokeUnlinkPublicIP(ctx context.Context, provider *client.NumSpotSDK, data *PublicIpModel) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	payload := numspot.UnlinkPublicIpJSONRequestBody{
		LinkPublicIpId: data.LinkPublicIP.ValueStringPointer(),
	}
	res, err := numspotClient.UnlinkPublicIpWithResponse(ctx, provider.SpaceID, data.Id.ValueString(), payload)
	if err != nil {
		return err
	}
	if res.StatusCode() != http.StatusOK {
		return utils.HandleError(res.Body)
	}

	return nil
}

func refreshState(ctx context.Context, provider *client.NumSpotSDK, id string, diags *diag.Diagnostics) *PublicIpModel {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return nil
	}

	// Refresh state
	res, err := numspotClient.ReadPublicIpsByIdWithResponse(ctx, provider.SpaceID, id)
	if err != nil {
		diags.AddError("Failed to read public ip", err.Error())
		return nil
	}

	if res.StatusCode() != http.StatusOK {
		apiError := utils.HandleError(res.Body)
		diags.AddError("Failed to read public ip", apiError.Error())
		return nil
	}

	tf := PublicIpFromHttpToTf(ctx, res.JSON200, diags)
	if diags.HasError() {
		return nil
	}

	return tf
}

func PublicIpsFromTfToAPIReadParams(ctx context.Context, tf PublicIpsDataSourceModel, diags *diag.Diagnostics) numspot.ReadPublicIpsParams {
	return numspot.ReadPublicIpsParams{
		LinkPublicIpIds: utils.TfStringListToStringPtrList(ctx, tf.LinkPublicIpIds, diags),
		NicIds:          utils.TfStringListToStringPtrList(ctx, tf.NicIds, diags),
		PrivateIps:      utils.TfStringListToStringPtrList(ctx, tf.PrivateIps, diags),
		TagKeys:         utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues:       utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
		Tags:            utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
		Ids:             utils.TfStringListToStringPtrList(ctx, tf.IDs, diags),
		VmIds:           utils.TfStringListToStringPtrList(ctx, tf.VmIds, diags),
	}
}

func PublicIpsFromHttpToTfDatasource(ctx context.Context, http *numspot.PublicIp, diags *diag.Diagnostics) *PublicIpModelDatasource {
	var tagsList types.List

	if http.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
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
	}
}
