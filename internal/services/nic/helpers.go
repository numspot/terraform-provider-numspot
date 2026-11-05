package nic

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services/nic/resource_nic"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

// securityGroupIdsFromGroups derives the flat security_group_ids list from the
// read response's SecurityGroups objects (the API exposes no top-level id list).
func securityGroupIDsFromGroups(ctx context.Context, groups *[]api.SecurityGroupLight, diags *diag.Diagnostics) types.List {
	ids := make([]string, 0)
	if groups != nil {
		for _, g := range *groups {
			if g.SecurityGroupId != nil {
				ids = append(ids, *g.SecurityGroupId)
			}
		}
	}
	return utils.StringListToTfListValue(ctx, ids, diags)
}

// lowerMacAddress normalizes the API MAC address (returned uppercased) to lowercase.
func lowerMacAddress(mac *string) types.String {
	if mac == nil {
		return types.StringNull()
	}
	return types.StringValue(strings.ToLower(*mac))
}

// linkNicBody builds the Link follow-up call body from the plan, or nil when no
// link_nic block is configured. Used by Create (extra arg) and Update.
func linkNicBody(tf *resource_nic.LinkNicValue) *api.LinkNicJSONRequestBody {
	if tf == nil || utils.IsTfValueNull(*tf) {
		return nil
	}
	return &api.LinkNicJSONRequestBody{
		DeviceNumber: utils.FromTfInt64ToInt(tf.DeviceNumber),
		VmId:         tf.VmId.ValueString(),
	}
}

// unlinkNicBody builds the Unlink call body from prior state, or nil when there
// was no link. Used by Delete (extra arg) and Update.
func unlinkNicBody(tf *resource_nic.LinkNicValue) *api.UnlinkNicJSONRequestBody {
	if tf == nil || utils.IsTfValueNull(*tf) {
		return nil
	}
	return &api.UnlinkNicJSONRequestBody{
		LinkNicId: tf.Id.ValueString(),
	}
}

// deserializeUpdateNicBody builds the attribute-update body (description + the
// link-nic update fields + security groups) from the plan.
func deserializeUpdateNicBody(ctx context.Context, tf *resource_nic.NicModel, diags *diag.Diagnostics) *api.UpdateNicJSONRequestBody {
	linkNic := api.LinkNicToUpdate{
		DeleteOnVmDeletion: utils.FromTfBoolToBoolPtr(tf.LinkNic.DeleteOnVmDeletion),
		LinkNicId:          utils.FromTfStringToStringPtr(tf.LinkNic.Id),
	}
	return &api.UpdateNicJSONRequestBody{
		SecurityGroupIds: utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds, diags),
		Description:      utils.FromTfStringToStringPtr(tf.Description),
		LinkNic:          &linkNic,
	}
}

// createPrivateIpsBody builds the create body's PrivateIpLight slice (a subset of
// the read PrivateIp type) from the plan.
func createPrivateIpsBody(ctx context.Context, set types.Set, diags *diag.Diagnostics) *[]api.PrivateIpLight {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}
	privateIps := utils.TfSetToGenericSet(func(a resource_nic.PrivateIpsValue) api.PrivateIpLight {
		return api.PrivateIpLight{
			IsPrimary: a.IsPrimary.ValueBoolPointer(),
			PrivateIp: a.PrivateIp.ValueStringPointer(),
		}
	}, ctx, set, diags)
	return &privateIps
}
