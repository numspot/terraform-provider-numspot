package snapshot

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type SnapshotsDataSourceModel struct {
	Items            []SnapshotModelDatasource `tfsdk:"items"`
	Descriptions     types.List                `tfsdk:"descriptions"`
	FromCreationDate types.String              `tfsdk:"from_creation_date"`
	Ids              types.List                `tfsdk:"ids"`
	IsPublic         types.Bool                `tfsdk:"is_public"`
	Progresses       types.List                `tfsdk:"progresses"`
	States           types.List                `tfsdk:"states"`
	TagKeys          types.List                `tfsdk:"tag_keys"`
	TagValues        types.List                `tfsdk:"tag_values"`
	Tags             types.List                `tfsdk:"tags"`
	ToCreationDate   types.String              `tfsdk:"to_creation_date"`
	VolumeIds        types.List                `tfsdk:"volume_ids"`
	VolumeSizes      types.List                `tfsdk:"volume_sizes"`
}

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
			"Unexpected Resource Configure Type",
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
	resp.Schema = SnapshotDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *snapshotsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan SnapshotsDataSourceModel
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

	objectItems := serializeSnapshots(ctx, numSpotSnapshot, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func deserializeSnapshotsParams(ctx context.Context, tf SnapshotsDataSourceModel, diags *diag.Diagnostics) numspot.ReadSnapshotsParams {
	return numspot.ReadSnapshotsParams{
		Descriptions:     utils.TfStringListToStringPtrList(ctx, tf.Descriptions, diags),
		FromCreationDate: utils.FromTfStringToStringPtr(tf.FromCreationDate),
		IsPublic:         utils.FromTfBoolToBoolPtr(tf.IsPublic),
		Progresses:       utils.TFInt64ListToIntListPointer(ctx, tf.Progresses, diags),
		States:           utils.TfStringListToStringPtrList(ctx, tf.States, diags),
		ToCreationDate:   utils.FromTfStringToStringPtr(tf.ToCreationDate),
		VolumeIds:        utils.TfStringListToStringPtrList(ctx, tf.VolumeIds, diags),
		VolumeSizes:      utils.TFInt64ListToIntListPointer(ctx, tf.VolumeSizes, diags),
		TagKeys:          utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues:        utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
		Tags:             utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
		Ids:              utils.TfStringListToStringPtrList(ctx, tf.Ids, diags),
	}
}

func serializeSnapshots(ctx context.Context, snapshots *[]numspot.Snapshot, diags *diag.Diagnostics) []SnapshotModelDatasource {
	return utils.FromHttpGenericListToTfList(ctx, snapshots, func(ctx context.Context, http *numspot.Snapshot, diags *diag.Diagnostics) *SnapshotModelDatasource {
		var (
			tagsList       types.List
			creationDateTf types.String
			progressTf     types.Int64
			volumeSizeTf   types.Int64
		)

		if http.Tags != nil {
			tagsList = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
			if diags.HasError() {
				return nil
			}
		}

		// Creation Date
		if http.CreationDate != nil {
			date := *http.CreationDate
			creationDateTf = types.StringValue(date.Format(time.RFC3339))
		}

		if http.Progress != nil {
			progress := int64(*http.Progress)
			progressTf = types.Int64PointerValue(&progress)
		}

		if http.VolumeSize != nil {
			volumeSize := int64(*http.VolumeSize)
			volumeSizeTf = types.Int64PointerValue(&volumeSize)
		}

		access, err := NewAccessValue(AccessValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"is_public": types.BoolPointerValue(http.Access.IsPublic),
			})
		if err != nil {
			return nil
		}

		return &SnapshotModelDatasource{
			Id:           types.StringPointerValue(http.Id),
			State:        types.StringPointerValue(http.State),
			Tags:         tagsList,
			CreationDate: creationDateTf,
			Description:  types.StringPointerValue(http.Description),
			VolumeId:     types.StringPointerValue(http.VolumeId),
			Progress:     progressTf,
			VolumeSize:   volumeSizeTf,
			Access:       access,
		}
	}, diags)
}
