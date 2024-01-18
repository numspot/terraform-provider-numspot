package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_key_pair"
)

func KeyPairFromTfToHttp(tf resource_key_pair.KeyPairModel) *api.KeypairSchema {
	return &api.KeypairSchema{
		Fingerprint: tf.Fingerprint.ValueStringPointer(),
		Name:        tf.Name.ValueStringPointer(),
	}
}

func KeyPairFromHttpToTf(http *api.KeypairSchema) resource_key_pair.KeyPairModel {
	return resource_key_pair.KeyPairModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Id:          types.StringPointerValue(http.Name),
		KeypairName: types.StringPointerValue(http.Name),
		Name:        types.StringPointerValue(http.Name),
		// PrivateKey:         types.StringPointerValue(htt.),
		PublicKey: types.StringValue(""),
	}
}

func KeyPairFromTfToCreateRequest(tf resource_key_pair.KeyPairModel) api.CreateKeypairJSONRequestBody {
	return api.CreateKeypairJSONRequestBody{
		Name:      tf.Name.ValueStringPointer(),
		PublicKey: tf.PublicKey.ValueStringPointer(),
	}
}
