package acl

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func CreateAclListFromTf(ctx context.Context, tf ACLsModel) ([]numspot.ACL, diag.Diagnostics) {
	if tf.ACLs.IsNull() || tf.ACLs.IsUnknown() {
		return nil, nil
	}
	acls := make([]ACLValue, 0, len(tf.ACLs.Elements()))
	diags := tf.ACLs.ElementsAs(ctx, &acls, false)

	aclsHttp := make([]numspot.ACL, 0, len(tf.ACLs.Elements()))

	for _, acl := range acls {
		permissionIdUuid, diags := utils.ParseUUID(acl.PermissionId.ValueString())
		if diags.HasError() {
			return nil, diags
		}

		aclsHttp = append(aclsHttp, numspot.ACL{
			PermissionId: permissionIdUuid,
			Resource:     tf.Resource.ValueString(),
			ResourceId:   acl.ResourceId.ValueString(),
			Service:      tf.Service.ValueString(),
			Subresource:  utils.FromTfStringToStringPtr(tf.Subresource),
		})
	}
	return aclsHttp, diags
}

func CreateTfAclFromHttp(ctx context.Context, http numspot.ACL) (ACLValue, diag.Diagnostics) {
	return NewACLValue(
		ACLValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"resource_id":   types.StringValue(http.ResourceId),
			"permission_id": types.StringValue(http.PermissionId.String()),
		},
	)
}
