package bucket

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/bucket/datasource_bucket"
)

var _ datasource.DataSource = &bucketsDataSource{}

type bucketsDataSource struct {
	services.BaseDataSource
}

// NewBucketsDataSource returns a new buckets data source.
func NewBucketsDataSource() datasource.DataSource {
	return &bucketsDataSource{}
}

// Metadata sets the type name.
func (d *bucketsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_buckets"
}

// Schema defines the schema.
func (d *bucketsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_bucket.BucketDataSourceSchema(ctx)
}

// Read reads .
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *bucketsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_bucket.BucketModel

	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	buckets, err := core.ReadBuckets(ctx, d.Provider)
	if err != nil {
		response.Diagnostics.AddError("unable to read buckets", err.Error())
		return
	}

	bucketItems := serializeBuckets(buckets, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = bucketItems.Items

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func serializeBuckets(data *core.ListBucketsOutput, diags *diag.Diagnostics) datasource_bucket.BucketModel {
	bucketType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":          types.StringType,
			"creation_date": types.StringType,
		},
	}

	var bucketElems []attr.Value
	for _, bucket := range data.AllBuckets.Buckets {
		date, _ := time.Parse(bucket.CreationDate, "2025-03-10T13:15:10.868Z")
		objValue, objDiag := types.ObjectValue(bucketType.AttrTypes, map[string]attr.Value{
			"name":          types.StringValue(bucket.Name),
			"creation_date": types.StringValue(date.Format(time.RFC3339)),
		})
		if objDiag.HasError() {
			diags.Append(objDiag...)
			continue
		}
		bucketElems = append(bucketElems, objValue)
	}

	bucketList, listDiag := types.ListValue(bucketType, bucketElems)
	if listDiag.HasError() {
		diags.Append(listDiag...)
	}

	return datasource_bucket.BucketModel{
		Items: bucketList,
	}
}
