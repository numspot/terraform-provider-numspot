package snapshot

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/snapshot/datasource_snapshot"
	"terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &snapshotsDataSource{}

// snapshotsDataSource is the Terraform data source implementation for numspot_snapshots.
type snapshotsDataSource struct {
	services.BaseDataSource
}

// NewSnapshotsDataSource returns a new snapshotsDataSource instance.
func NewSnapshotsDataSource() datasource.DataSource {
	return &snapshotsDataSource{}
}

// Metadata sets the data source type name.
func (d *snapshotsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snapshots"
}

// Schema defines the schema for the data source.
func (d *snapshotsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_snapshot.SnapshotDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *snapshotsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_snapshot.SnapshotModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	params := deserializeReadSnapshots(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	result, err := core.ReadSnapshotsWithParams(ctx, d.Provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read snapshots", err.Error())
		return
	}

	objectItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *result, mappingItemsValue)
	if serializeDiags.HasError() {
		response.Diagnostics.Append(serializeDiags...)
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, objectItems, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func mappingItemsValue(ctx context.Context, item *api.Snapshot) (datasource_snapshot.ItemsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	tagsList := types.ListNull(datasource_snapshot.ItemsValue{}.Type(ctx))

	if item.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_snapshot.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_snapshot.ItemsValue{}, serializeDiags
		}
	}

	value, valueDiags := datasource_snapshot.NewItemsValue(datasource_snapshot.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"access":        mappingAccess(ctx, item.Access, &diags),
		"creation_date": utils.TimePtrToTfString(item.CreationDate),
		"description":   types.StringValue(utils.ConvertStringPtrToString(item.Description)),
		"id":            types.StringValue(utils.ConvertStringPtrToString(item.Id)),
		"progress":      utils.FromIntPtrToTfInt64(item.Progress),
		"state":         types.StringValue(utils.ConvertStringPtrToString(item.State)),
		"volume_id":     types.StringValue(utils.ConvertStringPtrToString(item.VolumeId)),
		"volume_size":   utils.FromIntPtrToTfInt64(item.VolumeSize),
		"tags":          tagsList,
	})
	diags.Append(valueDiags...)
	return value, diags
}

func mappingTags(ctx context.Context, tag *api.ResourceTag) (datasource_snapshot.TagsValue, diag.Diagnostics) {
	return datasource_snapshot.NewTagsValue(datasource_snapshot.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}

func mappingAccess(ctx context.Context, item *api.Access, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(datasource_snapshot.AccessValue{}.AttributeTypes(ctx))
	}
	value, d := datasource_snapshot.NewAccessValue(datasource_snapshot.AccessValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"is_public": types.BoolPointerValue(item.IsPublic),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

func deserializeReadSnapshots(ctx context.Context, tf *datasource_snapshot.SnapshotModel, diags *diag.Diagnostics) *api.ReadSnapshotsParams {
	return &api.ReadSnapshotsParams{
		Descriptions:     utils.ConvertTfListToArrayOfString(ctx, tf.Descriptions, diags),
		FromCreationDate: tf.FromCreationDate.ValueStringPointer(),
		Progresses:       utils.ConvertTfListToArrayOfInt(ctx, tf.Progresses, diags),
		States:           utils.ConvertTfListToArrayOfString(ctx, tf.States, diags),
		TagKeys:          utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues:        utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
		Tags:             utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
		ToCreationDate:   tf.ToCreationDate.ValueStringPointer(),
		VolumeIds:        utils.ConvertTfListToArrayOfString(ctx, tf.VolumeIds, diags),
		VolumeSizes:      utils.ConvertTfListToArrayOfInt(ctx, tf.VolumeSizes, diags),
		IsPublic:         tf.IsPublic.ValueBoolPointer(),
		Ids:              utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
		Page: func() *api.PageParameter {
			if utils.IsTfValueNull(tf.Page) {
				return nil
			}
			return &api.PageParameter{NextToken: tf.Page.NextToken.ValueStringPointer(), Size: utils.FromTfInt64ToInt32Ptr(tf.Page.Size)}
		}(),
	}
}
