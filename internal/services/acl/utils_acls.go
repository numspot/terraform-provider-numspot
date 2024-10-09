package acl

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func CreateAclListFromTf(ctx context.Context, tf ACLsModel, diags *diag.Diagnostics) []numspot.ACL {
	if tf.ACLs.IsNull() || tf.ACLs.IsUnknown() {
		return nil
	}
	acls := make([]ACLValue, 0, len(tf.ACLs.Elements()))
	diags.Append(tf.ACLs.ElementsAs(ctx, &acls, false)...)

	aclsHttp := make([]numspot.ACL, 0, len(tf.ACLs.Elements()))

	for _, acl := range acls {
		permissionIdUuid := utils.ParseUUID(acl.PermissionId.ValueString(), diags)
		if diags.HasError() {
			return nil
		}

		aclsHttp = append(aclsHttp, numspot.ACL{
			PermissionId: permissionIdUuid,
			Resource:     tf.Resource.ValueString(),
			ResourceId:   acl.ResourceId.ValueString(),
			Service:      tf.Service.ValueString(),
			Subresource:  utils.FromTfStringToStringPtr(tf.Subresource),
		})
	}
	return aclsHttp
}

func CreateTfAclFromHttp(ctx context.Context, http numspot.ACL, diags *diag.Diagnostics) ACLValue {
	aclValue, diagnostics := NewACLValue(
		ACLValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"resource_id":   types.StringValue(http.ResourceId),
			"permission_id": types.StringValue(http.PermissionId.String()),
		},
	)
	diags.Append(diagnostics...)
	return aclValue
}
