package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_public_ip"
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

func PublicIpFromHttpToTf(elt *api.PublicIp, model *resource_public_ip.PublicIpModel) {
	model.Id = types.StringPointerValue(elt.Id)
	model.NicAccountId = types.StringPointerValue(elt.NicAccountId)
	model.NicId = types.StringPointerValue(elt.NicId)
	model.PrivateIp = types.StringPointerValue(elt.PrivateIp)
	model.PublicIp = types.StringPointerValue(elt.PublicIp)
	model.VmId = types.StringPointerValue(elt.VmId)
}

func invokeLinkPublicIP(ctx context.Context, provider Provider, data *resource_public_ip.PublicIpModel) (*string, error) {
	var payload api.LinkPublicIpJSONRequestBody
	if !data.VmId.IsNull() {
		payload = api.LinkPublicIpJSONRequestBody{VmId: data.VmId.ValueStringPointer()}
	} else {
		payload = api.LinkPublicIpJSONRequestBody{
			NicId:     data.NicId.ValueStringPointer(),
			PrivateIp: data.PrivateIp.ValueStringPointer(),
		}
	}
	res, err := provider.ApiClient.LinkPublicIpWithResponse(ctx, provider.SpaceID, data.Id.ValueString(), payload)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, utils.HandleError(res.Body)
	}

	return res.JSON200.LinkPublicIpId, nil
}

func invokeUnlinkPublicIP(ctx context.Context, provider Provider, data *resource_public_ip.PublicIpModel) error {
	payload := api.UnlinkPublicIpJSONRequestBody{}
	res, err := provider.ApiClient.UnlinkPublicIpWithResponse(ctx, provider.SpaceID, data.LinkPublicIP.ValueString(), payload)
	if err != nil {
		return err
	}
	if res.StatusCode() != http.StatusOK {
		return utils.HandleError(res.Body)
	}

	return nil
}

func refreshState(ctx context.Context, provider Provider, data *resource_public_ip.PublicIpModel) (*resource_public_ip.PublicIpModel, error) {
	// Refresh state
	res, err := provider.ApiClient.ReadPublicIpsByIdWithResponse(ctx, provider.SpaceID, data.Id.ValueString())
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		apiError := utils.HandleError(res.Body)
		return nil, apiError
	}

	PublicIpFromHttpToTf(res.JSON200, data)
	return data, nil
}
