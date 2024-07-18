package role

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"
)

func RolesFromTfToAPIReadParams(tf RolesDataSourceModel) numspot.ListRolesSpaceParams {
	return numspot.ListRolesSpaceParams{
		Name: tf.Name.ValueStringPointer(),
	}
}

func RegisteredRoleFromHTTPToTFDataSource(http numspot.RegisteredRole) RolesModel {
	return RolesModel{
		ID:          types.StringValue(http.Uuid.String()),
		Name:        types.StringValue(http.Name),
		Description: types.StringValue(http.Description),
		CreatedOn:   types.StringValue(http.CreatedOn.String()),
		UpdatedOn:   types.StringValue(http.UpdatedOn.String()),
	}
}
