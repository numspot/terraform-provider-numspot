package serviceaccount

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"
)

func ServiceAccountFromTFToCreateRequest(tf ServiceAccountModel) numspot.CreateServiceAccountSpaceJSONRequestBody {
	return numspot.CreateServiceAccountSpaceJSONRequestBody{
		Name: tf.Name.ValueString(),
	}
}

func CreateServiceAccountResponseFromHTTPToTF(ctx context.Context, http numspot.CreateServiceAccount201ResponseSchema) ServiceAccountModel {
	permissions := types.SetNull(types.StringType)
	roles := types.SetNull(types.StringType)

	return ServiceAccountModel{
		Id:                types.StringValue(http.Id),
		Name:              types.StringValue(http.Name),
		Secret:            types.StringValue(http.Secret),
		ServiceAccountId:  types.StringValue(http.Id),
		GlobalPermissions: permissions,
		Roles:             roles,
	}
}

func ServiceAccountEditedResponseFromHTTPToTF(ctx context.Context, http numspot.ServiceAccountEdited) ServiceAccountModel {
	permissions := types.SetNull(types.StringType)
	roles := types.SetNull(types.StringType)

	return ServiceAccountModel{
		Id:                types.StringValue(http.Id),
		Name:              types.StringValue(http.Name),
		ServiceAccountId:  types.StringValue(http.Id),
		GlobalPermissions: permissions,
		Roles:             roles,
	}
}

func ServiceAccountEditedResponseFromHTTPToTFDataSource(http numspot.ServiceAccountEdited) ServiceAccountModel {
	return ServiceAccountModel{
		Id:   types.StringValue(http.Id),
		Name: types.StringValue(http.Name),
	}
}

func ServiceAccountsFromTfToAPIReadParams(tf ServiceAccountsDataSourceModel) numspot.ListServiceAccountSpaceParams {
	return numspot.ListServiceAccountSpaceParams{}
}
