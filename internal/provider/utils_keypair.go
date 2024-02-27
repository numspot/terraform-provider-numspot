package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_key_pair"
)

func KeyPairFromTfToHttp(tf *resource_key_pair.KeyPairModel) *api.Keypair {
	return &api.Keypair{
		Fingerprint: tf.Fingerprint.ValueStringPointer(),
		Name:        tf.Name.ValueStringPointer(),
	}
}

func KeyPairFromCreateHttpToTf(http *api.CreateKeypairResponseSchema) resource_key_pair.KeyPairModel {
	res := resource_key_pair.KeyPairModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Id:          types.StringPointerValue(http.Name),
		Name:        types.StringPointerValue(http.Name),
		PrivateKey:  types.StringPointerValue(http.PrivateKey),
	}

	return res
}

func KeyPairFromReadHttpToTf(http *api.ReadKeypairsByIdResponseSchema) resource_key_pair.KeyPairModel {
	res := resource_key_pair.KeyPairModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Id:          types.StringPointerValue(http.Name),
		Name:        types.StringPointerValue(http.Name),
	}

	return res
}

func KeyPairFromImportHttpToTf(http *api.CreateKeypairResponseSchema) resource_key_pair.KeyPairModel {
	res := resource_key_pair.KeyPairModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Id:          types.StringPointerValue(http.Name),
		Name:        types.StringPointerValue(http.Name),
		PrivateKey:  types.StringPointerValue(http.PrivateKey),
	}

	return res
}

func KeyPairFromTfToCreateRequest(tf *resource_key_pair.KeyPairModel) api.CreateKeypairJSONRequestBody {
	return api.CreateKeypairJSONRequestBody{
		Name:      tf.Name.ValueString(),
		PublicKey: tf.PublicKey.ValueStringPointer(),
	}
}
