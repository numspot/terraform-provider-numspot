package snapshot

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services/snapshot/datasource_snapshot"
	"terraform-provider-numspot/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &snapshotsDataSource{}
)

func (d *snapshotsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(*client.NumSpotSDK)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Datasource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	d.provider = provider
}

func NewSnapshotsDataSource() datasource.DataSource {
	return &snapshotsDataSource{}
}

type snapshotsDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *snapshotsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snapshots"
}

// Schema defines the schema for the data source.
func (d *snapshotsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_snapshot.SnapshotDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *snapshotsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_snapshot.SnapshotModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := deserializeSnapshotsParams(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	numSpotSnapshot, err := core.ReadSnapshotsWithParams(ctx, d.provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read snapshot", err.Error())
		return
	}

	objectItems := utils.SerializeDatasourceItemsWithDiags(ctx, *numSpotSnapshot, &response.Diagnostics, mappingItemsValue)
	if response.Diagnostics.HasError() {
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, objectItems, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func deserializeSnapshotsParams(ctx context.Context, tf datasource_snapshot.SnapshotModel, diags *diag.Diagnostics) api.ReadSnapshotsParams {
	return api.ReadSnapshotsParams{
		Descriptions:     utils.ConvertTfListToArrayOfString(ctx, tf.Descriptions, diags),
		FromCreationDate: utils.FromTfStringToStringPtr(tf.FromCreationDate),
		IsPublic:         utils.FromTfBoolToBoolPtr(tf.IsPublic),
		Progresses:       utils.ConvertTfListToArrayOfInt(ctx, tf.Progresses, diags),
		States:           utils.ConvertTfListToArrayOfString(ctx, tf.States, diags),
		ToCreationDate:   utils.FromTfStringToStringPtr(tf.ToCreationDate),
		VolumeIds:        utils.ConvertTfListToArrayOfString(ctx, tf.VolumeIds, diags),
		VolumeSizes:      utils.ConvertTfListToArrayOfInt(ctx, tf.VolumeSizes, diags),
		TagKeys:          utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues:        utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
		Tags:             utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
		Ids:              utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
	}
}

func mappingItemsValue(ctx context.Context, snapshot api.Snapshot, diags *diag.Diagnostics) (datasource_snapshot.ItemsValue, diag.Diagnostics) {
	var serializeDiags diag.Diagnostics

	tagsList := types.ListNull(datasource_snapshot.ItemsValue{}.Type(ctx))
	accessObject := basetypes.ObjectValue{}
	creationDateTf := basetypes.StringValue{}
	progressTf := basetypes.Int64Value{}
	volumeSizeTf := basetypes.Int64Value{}

	if snapshot.Access != nil {
		accessObject, serializeDiags = mappingAccess(ctx, snapshot, diags)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	}

	if snapshot.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *snapshot.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_snapshot.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_snapshot.ItemsValue{}, serializeDiags
		}
	}

	if snapshot.CreationDate != nil {
		date := *snapshot.CreationDate
		creationDateTf = types.StringValue(date.Format(time.RFC3339))
	}

	if snapshot.Progress != nil {
		progress := int64(*snapshot.Progress)
		progressTf = types.Int64PointerValue(&progress)
	}

	if snapshot.VolumeSize != nil {
		volumeSize := int64(*snapshot.VolumeSize)
		volumeSizeTf = types.Int64PointerValue(&volumeSize)
	}

	return datasource_snapshot.NewItemsValue(datasource_snapshot.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"access":        accessObject,
		"creation_date": creationDateTf,
		"description":   types.StringValue(utils.ConvertStringPtrToString(snapshot.Description)),
		"id":            types.StringValue(utils.ConvertStringPtrToString(snapshot.Id)),
		"progress":      progressTf,
		"state":         types.StringValue(utils.ConvertStringPtrToString(snapshot.State)),
		"tags":          tagsList,
		"volume_id":     types.StringValue(utils.ConvertStringPtrToString(snapshot.VolumeId)),
		"volume_size":   volumeSizeTf,
	})
}

func mappingTags(ctx context.Context, tag api.ResourceTag) (datasource_snapshot.TagsValue, diag.Diagnostics) {
	return datasource_snapshot.NewTagsValue(datasource_snapshot.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}

func mappingAccess(ctx context.Context, snapshot api.Snapshot, diags *diag.Diagnostics) (basetypes.ObjectValue, diag.Diagnostics) {
	var mappingDiags diag.Diagnostics
	var accessValue datasource_snapshot.AccessValue

	accessValue, mappingDiags = datasource_snapshot.NewAccessValue(datasource_snapshot.AccessValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"is_public": types.BoolPointerValue(snapshot.Access.IsPublic),
		})
	if mappingDiags.HasError() {
		diags.Append(mappingDiags...)
	}

	accessObject, mappingDiags := accessValue.ToObjectValue(ctx)
	if mappingDiags.HasError() {
		diags.Append(mappingDiags...)
	}

	return accessObject, mappingDiags
}
