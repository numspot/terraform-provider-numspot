package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_service_account"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_service_account"
)

func ServiceAccountFromTFToCreateRequest(tf resource_service_account.ServiceAccountModel) numspot.CreateServiceAccountSpaceJSONRequestBody {
	return numspot.CreateServiceAccountSpaceJSONRequestBody{
		Name: tf.Name.ValueString(),
	}
}

func CreateServiceAccountResponseFromHTTPToTF(ctx context.Context, http numspot.CreateServiceAccount201ResponseSchema) resource_service_account.ServiceAccountModel {
	permissions := types.SetNull(types.StringType)
	roles := types.SetNull(types.StringType)

	return resource_service_account.ServiceAccountModel{
		Id:                types.StringValue(http.Id),
		Name:              types.StringValue(http.Name),
		Secret:            types.StringValue(http.Secret),
		ServiceAccountId:  types.StringValue(http.Id),
		GlobalPermissions: permissions,
		Roles:             roles,
	}
}

func ServiceAccountEditedResponseFromHTTPToTF(ctx context.Context, http numspot.ServiceAccountEdited) resource_service_account.ServiceAccountModel {
	permissions := types.SetNull(types.StringType)
	roles := types.SetNull(types.StringType)

	return resource_service_account.ServiceAccountModel{
		Id:                types.StringValue(http.Id),
		Name:              types.StringValue(http.Name),
		ServiceAccountId:  types.StringValue(http.Id),
		GlobalPermissions: permissions,
		Roles:             roles,
	}
}

func ServiceAccountEditedResponseFromHTTPToTFDataSource(http numspot.ServiceAccountEdited) datasource_service_account.ServiceAccountModel {
	return datasource_service_account.ServiceAccountModel{
		ID:   types.StringValue(http.Id),
		Name: types.StringValue(http.Name),
	}
}

func ServiceAccountsFromTfToAPIReadParams(tf ServiceAccountsDataSourceModel) numspot.ListServiceAccountSpaceParams {
	return numspot.ListServiceAccountSpaceParams{}
}
