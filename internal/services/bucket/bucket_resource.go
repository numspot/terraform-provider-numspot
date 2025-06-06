package bucket

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/bucket/resource_bucket"
)

var (
	_ resource.Resource                = &bucketResource{}
	_ resource.ResourceWithConfigure   = &bucketResource{}
	_ resource.ResourceWithImportState = &bucketResource{}
)

type bucketResource struct {
	provider *client.NumSpotSDK
}

func NewBucketResource() resource.Resource {
	return &bucketResource{}
}

func (r *bucketResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *bucketResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), request, response)
}

func (r *bucketResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_bucket"
}

func (r *bucketResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_bucket.BucketResourceSchema(ctx)
}

func (r *bucketResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.Root("name").Expression(),
		),
	}
}

func (r *bucketResource) Create(ctx context.Context, req resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_bucket.BucketModel

	response.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	bucketName := plan.Name.ValueString()

	err := core.CreateBucket(ctx, r.provider, bucketName)
	if err != nil {
		response.Diagnostics.AddError("unable to create bucket", err.Error())
		return
	}

	state := serializeNumSpotCreateBucket(bucketName)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *bucketResource) Read(ctx context.Context, req resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_bucket.BucketModel

	response.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	bucketName := state.Name.ValueString()

	bucket, err := core.ReadBucket(ctx, r.provider, bucketName)
	if err != nil {
		response.Diagnostics.AddError("unable to read bucket", err.Error())
		return
	}

	newState := serializeNumSpotBucket(*bucket)
	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *bucketResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
}

func (r *bucketResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_bucket.BucketModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteBucket(ctx, r.provider, state.Name.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete bucket", err.Error())
		return
	}
}

func serializeNumSpotCreateBucket(bucketName string) resource_bucket.BucketModel {
	return resource_bucket.BucketModel{
		Name: types.StringPointerValue(&bucketName),
	}
}

func serializeNumSpotBucket(bucket core.Bucket) resource_bucket.BucketModel {
	return resource_bucket.BucketModel{
		Name: types.StringValue(bucket.Name),
	}
}
