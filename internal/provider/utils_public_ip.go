package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_public_ip"
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

func PublicIpFromHttpToTf(elt *iaas.PublicIp) resource_public_ip.PublicIpModel {
	return resource_public_ip.PublicIpModel{
		Id:           types.StringPointerValue(elt.Id),
		NicAccountId: types.StringPointerValue(elt.NicAccountId),
		NicId:        types.StringPointerValue(elt.NicId),
		PrivateIp:    types.StringPointerValue(elt.PrivateIp),
		PublicIp:     types.StringPointerValue(elt.PublicIp),
		VmId:         types.StringPointerValue(elt.VmId),
		LinkPublicIP: types.StringPointerValue(elt.LinkPublicIpId),
	}
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
	payload := iaas.UnlinkPublicIpJSONRequestBody{
		LinkPublicIpId: data.LinkPublicIP.ValueStringPointer(),
	}
	res, err := provider.ApiClient.UnlinkPublicIpWithResponse(ctx, provider.SpaceID, data.LinkPublicIP.ValueString(), payload)
	if err != nil {
		return err
	}
	if res.StatusCode() != http.StatusOK {
		return utils.HandleError(res.Body)
	}

	return nil
}

func refreshState(ctx context.Context, provider Provider, id string) (*resource_public_ip.PublicIpModel, error) {
	// Refresh state
	res, err := provider.ApiClient.ReadPublicIpsByIdWithResponse(ctx, provider.SpaceID, id)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		apiError := utils.HandleError(res.Body)
		return nil, apiError
	}

	tf := PublicIpFromHttpToTf(res.JSON200)
	return &tf, nil
}
