package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iam"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_service_account"
)

func ServiceAccountFromTFToCreateRequest(tf resource_service_account.ServiceAccountModel) iam.CreateServiceAccountSpaceJSONRequestBody {
	return iam.CreateServiceAccountSpaceJSONRequestBody{
		Name: tf.Name.ValueString(),
	}
}

func CreateServiceAccountResponseFromHTTPToTF(http iam.CreateServiceAccountResponseSchema) resource_service_account.ServiceAccountModel {
	permissions := types.ListNull(types.StringType)

	return resource_service_account.ServiceAccountModel{
		Id:                types.StringValue(http.Id),
		Name:              types.StringValue(http.Name),
		Secret:            types.StringValue(http.Secret),
		ServiceAccountId:  types.StringValue(http.Id),
		GlobalPermissions: permissions,
	}
}

func ServiceAccountEditedResponseFromHTTPToTF(http iam.ServiceAccountEdited) resource_service_account.ServiceAccountModel {
	permissions := types.ListNull(types.StringType)

	return resource_service_account.ServiceAccountModel{
		Id:                types.StringValue(http.Id),
		Name:              types.StringValue(http.Name),
		ServiceAccountId:  types.StringValue(http.Id),
		GlobalPermissions: permissions,
	}
}
