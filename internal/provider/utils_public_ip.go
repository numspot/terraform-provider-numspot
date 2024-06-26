package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_public_ip"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_public_ip"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var errNicVmConflict = fmt.Errorf("couldn't have nicID and vmID at the same time")

type PublicIPChangeSet struct {
	Unlink bool
	Link   bool
	Err    error
}

func ComputePublicIPChangeSet(plan, state *resource_public_ip.PublicIpModel) PublicIPChangeSet {
	c := PublicIPChangeSet{Err: nil}
	c.Unlink = plan.VmId.IsNull() && !state.VmId.IsNull() ||
		plan.NicId.IsUnknown() && !state.NicId.IsNull()

	switch {
	case plan.NicId.IsUnknown() && plan.VmId.IsNull():
		c.Link = false
	case !plan.NicId.IsUnknown() && !plan.VmId.IsNull():
		c.Err = errNicVmConflict
	case !plan.NicId.IsUnknown():
		c.Link = true
	case !plan.VmId.IsNull():
		c.Link = true
	}
	return c
}

func PublicIpFromHttpToTf(ctx context.Context, elt *iaas.PublicIp) (*resource_public_ip.PublicIpModel, diag.Diagnostics) {
	var (
		diags    diag.Diagnostics
		tagsList types.List
	)

	if elt.Tags != nil {
		tagsList, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *elt.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &resource_public_ip.PublicIpModel{
		Id:           types.StringPointerValue(elt.Id),
		NicId:        types.StringPointerValue(elt.NicId),
		PrivateIp:    types.StringPointerValue(elt.PrivateIp),
		PublicIp:     types.StringPointerValue(elt.PublicIp),
		VmId:         types.StringPointerValue(elt.VmId),
		LinkPublicIP: types.StringPointerValue(elt.LinkPublicIpId),
		Tags:         tagsList,
	}, nil
}

func invokeLinkPublicIP(ctx context.Context, provider Provider, data *resource_public_ip.PublicIpModel) (*string, error) {
	var payload iaas.LinkPublicIpJSONRequestBody
	if !data.VmId.IsNull() {
		payload = iaas.LinkPublicIpJSONRequestBody{VmId: data.VmId.ValueStringPointer()}
	} else {
		payload = iaas.LinkPublicIpJSONRequestBody{
			NicId:     data.NicId.ValueStringPointer(),
			PrivateIp: data.PrivateIp.ValueStringPointer(),
		}
	}
	res, err := provider.IaasClient.LinkPublicIpWithResponse(ctx, provider.SpaceID, data.Id.ValueString(), payload)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, utils.HandleError(res.Body)
	}

	return res.JSON200.LinkPublicIpId, nil
}

func invokeUnlinkPublicIP(ctx context.Context, provider Provider, data *resource_public_ip.PublicIpModel) error {
	payload := iaas.UnlinkPublicIpJSONRequestBody{
		LinkPublicIpId: data.LinkPublicIP.ValueStringPointer(),
	}
	res, err := provider.IaasClient.UnlinkPublicIpWithResponse(ctx, provider.SpaceID, data.Id.ValueString(), payload)
	if err != nil {
		return err
	}
	if res.StatusCode() != http.StatusOK {
		return utils.HandleError(res.Body)
	}

	return nil
}

func refreshState(ctx context.Context, provider Provider, id string) (*resource_public_ip.PublicIpModel, diag.Diagnostics) {
	// Refresh state
	var diags diag.Diagnostics

	res, err := provider.IaasClient.ReadPublicIpsByIdWithResponse(ctx, provider.SpaceID, id)
	if err != nil {
		diags.AddError("Failed to read public ip", err.Error())
		return nil, diags
	}

	if res.StatusCode() != http.StatusOK {
		apiError := utils.HandleError(res.Body)
		diags.AddError("Failed to read public ip", apiError.Error())
		return nil, diags
	}

	tf, diags := PublicIpFromHttpToTf(ctx, res.JSON200)
	if diags.HasError() {
		return nil, diags
	}

	return tf, diags
}

func PublicIpsFromTfToAPIReadParams(ctx context.Context, tf PublicIpsDataSourceModel) iaas.ReadPublicIpsParams {
	return iaas.ReadPublicIpsParams{
		LinkPublicIpIds: utils.TfStringListToStringPtrList(ctx, tf.LinkPublicIpIds),
		NicIds:          utils.TfStringListToStringPtrList(ctx, tf.NicIds),
		PrivateIps:      utils.TfStringListToStringPtrList(ctx, tf.PrivateIps),
		TagKeys:         utils.TfStringListToStringPtrList(ctx, tf.TagKeys),
		TagValues:       utils.TfStringListToStringPtrList(ctx, tf.TagValues),
		Tags:            utils.TfStringListToStringPtrList(ctx, tf.Tags),
		Ids:             utils.TfStringListToStringPtrList(ctx, tf.IDs),
		VmIds:           utils.TfStringListToStringPtrList(ctx, tf.VmIds),
	}
}

func PublicIpsFromHttpToTfDatasource(ctx context.Context, http *iaas.PublicIp) (*datasource_public_ip.PublicIpModel, diag.Diagnostics) {
	var (
		tagsList types.List
		diag     diag.Diagnostics
	)

	if http.Tags != nil {
		tagsList, diag = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diag.HasError() {
			return nil, diag
		}
	}

	return &datasource_public_ip.PublicIpModel{
		Id:             types.StringPointerValue(http.Id),
		NicId:          types.StringPointerValue(http.NicId),
		PrivateIp:      types.StringPointerValue(http.PrivateIp),
		PublicIp:       types.StringPointerValue(http.PublicIp),
		VmId:           types.StringPointerValue(http.VmId),
		LinkPublicIpId: types.StringPointerValue(http.LinkPublicIpId),
		Tags:           tagsList,
	}, nil
}
