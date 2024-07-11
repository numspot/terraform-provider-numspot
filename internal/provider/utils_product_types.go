package provider

/*

Product Types are not handled for now


import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_product_type"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func ProductTypesFromTfToAPIReadParams(ctx context.Context, tf ProductTypesDataSourceModel) numspot.ReadProductTypesParams {
	return numspot.ReadProductTypesParams{
		Ids: utils.TfStringListToStringPtrList(ctx, tf.IDs),
	}
}

func ProductTypesFromHttpToTfDatasource(ctx context.Context, http *numspot.ProductType) (*datasource_product_type.ProductTypeModel, diag.Diagnostics) {

	return &datasource_product_type.ProductTypeModel{
		Id:          types.StringPointerValue(http.Id),
		Description: types.StringPointerValue(http.Description),
		Vendor:      types.StringPointerValue(http.Vendor),
	}, nil
}
*/
