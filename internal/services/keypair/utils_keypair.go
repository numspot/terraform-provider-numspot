package keypair

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func KeyPairFromTfToHttp(tf *KeyPairModel) *numspot.Keypair {
	return &numspot.Keypair{
		Fingerprint: tf.Fingerprint.ValueStringPointer(),
		Name:        tf.Name.ValueStringPointer(),
	}
}

func KeyPairFromCreateHttpToTf(http *numspot.CreateKeypairResponseSchema) KeyPairModel {
	res := KeyPairModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Id:          types.StringPointerValue(http.Name),
		Name:        types.StringPointerValue(http.Name),
		PrivateKey:  types.StringPointerValue(http.PrivateKey),
	}

	return res
}

func KeyPairFromReadHttpToTf(http *numspot.ReadKeypairsByIdResponseSchema) KeyPairModel {
	res := KeyPairModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Id:          types.StringPointerValue(http.Name),
		Name:        types.StringPointerValue(http.Name),
	}

	return res
}

func KeyPairFromImportHttpToTf(http *numspot.CreateKeypairResponseSchema) KeyPairModel {
	res := KeyPairModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Id:          types.StringPointerValue(http.Name),
		Name:        types.StringPointerValue(http.Name),
		PrivateKey:  types.StringPointerValue(http.PrivateKey),
	}

	return res
}

func KeyPairFromTfToCreateRequest(tf *KeyPairModel) numspot.CreateKeypairJSONRequestBody {
	return numspot.CreateKeypairJSONRequestBody{
		Name:      tf.Name.ValueString(),
		PublicKey: tf.PublicKey.ValueStringPointer(),
	}
}

func KeypairsFromTfToAPIReadParams(ctx context.Context, tf KeypairsDataSourceModel, diags *diag.Diagnostics) numspot.ReadKeypairsParams {
	return numspot.ReadKeypairsParams{
		KeypairFingerprints: utils.TfStringListToStringPtrList(ctx, tf.KeypairFingerprints, diags),
		KeypairNames:        utils.TfStringListToStringPtrList(ctx, tf.KeypairNames, diags),
		KeypairTypes:        utils.TfStringListToStringPtrList(ctx, tf.KeypairTypes, diags),
	}
}

func KeypairsFromHttpToTfDatasource(ctx context.Context, http *numspot.Keypair, diags *diag.Diagnostics) *KeyPairDatasourceItemModel {
	return &KeyPairDatasourceItemModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Name:        types.StringPointerValue(http.Name),
		Type:        types.StringPointerValue(http.Type),
	}
}
