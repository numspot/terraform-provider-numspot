package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iam"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_permissions"
)

func RegisteredPermissionFromHTTPToTFDataSource(http iam.RegisteredPermission) datasource_permissions.PermissionModel {
	return datasource_permissions.PermissionModel{
		ID:          types.StringValue(http.Uuid.String()),
		Name:        types.StringValue(http.Name),
		Description: types.StringValue(http.Description),
		Service:     types.StringValue(http.Service),
		Resource:    types.StringPointerValue(http.Resource),
		Subresource: types.StringPointerValue(http.SubResource),
		Action:      types.StringValue(http.Action),
		CreatedOn:   types.StringValue(http.CreatedOn.String()),
		UpdatedOn:   types.StringValue(http.UpdatedOn.String()),
	}
}

func PermissionsFromTfToAPIReadParams(tf PermissionsDataSourceModel) iam.ListPermissionsSpaceParams {
	return iam.ListPermissionsSpaceParams{
		Service:     utils.FromTfStringToStringPtr(tf.Service),
		Resource:    utils.FromTfStringToStringPtr(tf.Resource),
		Subresource: utils.FromTfStringToStringPtr(tf.Subresource),
		Action:      utils.FromTfStringToStringPtr(tf.Action),
	}
}
