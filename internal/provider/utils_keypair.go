package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_key_pair"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_key_pair"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func KeyPairFromTfToHttp(tf *resource_key_pair.KeyPairModel) *numspot.Keypair {
	return &numspot.Keypair{
		Fingerprint: tf.Fingerprint.ValueStringPointer(),
		Name:        tf.Name.ValueStringPointer(),
	}
}

func KeyPairFromCreateHttpToTf(http *numspot.CreateKeypairResponseSchema) resource_key_pair.KeyPairModel {
	res := resource_key_pair.KeyPairModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Id:          types.StringPointerValue(http.Name),
		Name:        types.StringPointerValue(http.Name),
		PrivateKey:  types.StringPointerValue(http.PrivateKey),
	}

	return res
}

func KeyPairFromReadHttpToTf(http *numspot.ReadKeypairsByIdResponseSchema) resource_key_pair.KeyPairModel {
	res := resource_key_pair.KeyPairModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Id:          types.StringPointerValue(http.Name),
		Name:        types.StringPointerValue(http.Name),
	}

	return res
}

func KeyPairFromImportHttpToTf(http *numspot.CreateKeypairResponseSchema) resource_key_pair.KeyPairModel {
	res := resource_key_pair.KeyPairModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Id:          types.StringPointerValue(http.Name),
		Name:        types.StringPointerValue(http.Name),
		PrivateKey:  types.StringPointerValue(http.PrivateKey),
	}

	return res
}

func KeyPairFromTfToCreateRequest(tf *resource_key_pair.KeyPairModel) numspot.CreateKeypairJSONRequestBody {
	return numspot.CreateKeypairJSONRequestBody{
		Name:      tf.Name.ValueString(),
		PublicKey: tf.PublicKey.ValueStringPointer(),
	}
}

func KeypairsFromTfToAPIReadParams(ctx context.Context, tf KeypairsDataSourceModel) numspot.ReadKeypairsParams {
	return numspot.ReadKeypairsParams{
		KeypairFingerprints: utils.TfStringListToStringPtrList(ctx, tf.Fingerprints),
		KeypairNames:        utils.TfStringListToStringPtrList(ctx, tf.Names),
		KeypairTypes:        utils.TfStringListToStringPtrList(ctx, tf.Types),
	}
}

func KeypairsFromHttpToTfDatasource(ctx context.Context, http *numspot.Keypair) (*datasource_key_pair.KeyPairModel, diag.Diagnostics) {
	return &datasource_key_pair.KeyPairModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Name:        types.StringPointerValue(http.Name),
		Type:        types.StringPointerValue(http.Type),
	}, nil
}
