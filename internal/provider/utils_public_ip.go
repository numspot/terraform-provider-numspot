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

func PublicIpFromTfToHttp(tf *resource_public_ip.PublicIpModel) *api.PublicIpSchema {
	return &api.PublicIpSchema{
		Id:           nil,
		NicAccountId: nil,
		NicId:        nil,
		PrivateIp:    nil,
		PublicIp:     nil,
		Tags:         nil,
		VmId:         nil,
	}
}

func PublicIpFromHttpToTf(http *api.PublicIpSchema, model *resource_public_ip.PublicIpModel) {
	model.Id = types.StringPointerValue(http.Id)
	model.NicAccountId = types.StringPointerValue(http.NicAccountId)
	model.NicId = types.StringPointerValue(http.NicId)
	model.PrivateIp = types.StringPointerValue(http.PrivateIp)
	model.PublicIp = types.StringPointerValue(http.PublicIp)
	model.VmId = types.StringPointerValue(http.VmId)
}

func invokeLinkPublicIP(ctx context.Context, client *api.ClientWithResponses, data *resource_public_ip.PublicIpModel) (*string, error) {
	var payload api.LinkPublicIpJSONRequestBody
	if !data.VmId.IsNull() {
		payload = api.LinkPublicIpJSONRequestBody{VmId: data.VmId.ValueStringPointer()}
	} else {
		payload = api.LinkPublicIpJSONRequestBody{
			NicId:     data.NicId.ValueStringPointer(),
			PrivateIp: data.PrivateIp.ValueStringPointer(),
		}
	}
	res, err := client.LinkPublicIpWithResponse(ctx, data.Id.ValueString(), payload)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, utils.HandleError(res.Body)
	}

	return res.JSON200.Id, nil
}

func invokeUnlinkPublicIP(ctx context.Context, client *api.ClientWithResponses, data *resource_public_ip.PublicIpModel) error {
	payload := api.UnlinkPublicIpJSONRequestBody{}
	res, err := client.UnlinkPublicIpWithResponse(ctx, data.LinkPublicIP.ValueString(), payload)
	if err != nil {
		return err
	}
	if res.StatusCode() != http.StatusOK {
		return utils.HandleError(res.Body)
	}

	return nil
}

func refreshState(ctx context.Context, client *api.ClientWithResponses, data *resource_public_ip.PublicIpModel) (*resource_public_ip.PublicIpModel, error) {
	// Refresh state
	res, err := client.ReadPublicIpsByIdWithResponse(ctx, data.Id.ValueString())
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
