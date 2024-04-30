package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_key_pair"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_key_pair"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func KeyPairFromTfToHttp(tf *resource_key_pair.KeyPairModel) *iaas.Keypair {
	return &iaas.Keypair{
		Fingerprint: tf.Fingerprint.ValueStringPointer(),
		Name:        tf.Name.ValueStringPointer(),
	}
}

func KeyPairFromCreateHttpToTf(http *iaas.CreateKeypairResponseSchema) resource_key_pair.KeyPairModel {
	res := resource_key_pair.KeyPairModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Id:          types.StringPointerValue(http.Name),
		Name:        types.StringPointerValue(http.Name),
		PrivateKey:  types.StringPointerValue(http.PrivateKey),
	}

	return res
}

func KeyPairFromReadHttpToTf(http *iaas.ReadKeypairsByIdResponseSchema) resource_key_pair.KeyPairModel {
	res := resource_key_pair.KeyPairModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Id:          types.StringPointerValue(http.Name),
		Name:        types.StringPointerValue(http.Name),
	}

	return res
}

func KeyPairFromImportHttpToTf(http *iaas.CreateKeypairResponseSchema) resource_key_pair.KeyPairModel {
	res := resource_key_pair.KeyPairModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Id:          types.StringPointerValue(http.Name),
		Name:        types.StringPointerValue(http.Name),
		PrivateKey:  types.StringPointerValue(http.PrivateKey),
	}

	return res
}

func KeyPairFromTfToCreateRequest(tf *resource_key_pair.KeyPairModel) iaas.CreateKeypairJSONRequestBody {
	return iaas.CreateKeypairJSONRequestBody{
		Name:      tf.Name.ValueString(),
		PublicKey: tf.PublicKey.ValueStringPointer(),
	}
}

func KeypairsFromTfToAPIReadParams(ctx context.Context, tf KeypairsDataSourceModel) iaas.ReadKeypairsParams {
	return iaas.ReadKeypairsParams{
		KeypairFingerprints: utils.TfStringListToStringPtrList(ctx, tf.Fingerprints),
		KeypairNames:        utils.TfStringListToStringPtrList(ctx, tf.Names),
		KeypairTypes:        utils.TfStringListToStringPtrList(ctx, tf.Types),
	}
}

func KeypairsFromHttpToTfDatasource(ctx context.Context, http *iaas.Keypair) *datasource_key_pair.KeyPairModel {
	return &datasource_key_pair.KeyPairModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Name:        types.StringPointerValue(http.Name),
		Type:        types.StringPointerValue(http.Type),
	}
}
