package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_roles"
)

func RolesFromTfToAPIReadParams(tf RolesDataSourceModel) numspot.ListRolesSpaceParams {
	return numspot.ListRolesSpaceParams{
		Name: tf.Name.ValueStringPointer(),
	}
}

func RegisteredRoleFromHTTPToTFDataSource(http numspot.RegisteredRole) datasource_roles.RolesModel {
	return datasource_roles.RolesModel{
		ID:          types.StringValue(http.Uuid.String()),
		Name:        types.StringValue(http.Name),
		Description: types.StringValue(http.Description),
		CreatedOn:   types.StringValue(http.CreatedOn.String()),
		UpdatedOn:   types.StringValue(http.UpdatedOn.String()),
	}
}
