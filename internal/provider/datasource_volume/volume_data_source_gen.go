// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package datasource_volume

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func VolumeDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"items": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"availability_zone_name": schema.StringAttribute{
							Computed:            true,
							Description:         "The Subregion in which the volume was created.",
							MarkdownDescription: "The Subregion in which the volume was created.",
						},
						"creation_date": schema.StringAttribute{
							Computed:            true,
							Description:         "The date and time of creation of the volume.",
							MarkdownDescription: "The date and time of creation of the volume.",
						},
						"id": schema.StringAttribute{
							Required:            true,
							Description:         "ID for ReadVolumes",
							MarkdownDescription: "ID for ReadVolumes",
						},
						"iops": schema.Int64Attribute{
							Computed:            true,
							Description:         "The number of I/O operations per second (IOPS):<br />\n- For `io1` volumes, the number of provisioned IOPS<br />\n- For `gp2` volumes, the baseline performance of the volume",
							MarkdownDescription: "The number of I/O operations per second (IOPS):<br />\n- For `io1` volumes, the number of provisioned IOPS<br />\n- For `gp2` volumes, the baseline performance of the volume",
						},
						"linked_volumes": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"delete_on_vm_deletion": schema.BoolAttribute{
										Computed:            true,
										Description:         "If true, the volume is deleted when terminating the VM. If false, the volume is not deleted when terminating the VM.",
										MarkdownDescription: "If true, the volume is deleted when terminating the VM. If false, the volume is not deleted when terminating the VM.",
									},
									"device_name": schema.StringAttribute{
										Computed:            true,
										Description:         "The name of the device.",
										MarkdownDescription: "The name of the device.",
									},
									"id": schema.StringAttribute{
										Computed:            true,
										Description:         "The ID of the volume.",
										MarkdownDescription: "The ID of the volume.",
									},
									"state": schema.StringAttribute{
										Computed:            true,
										Description:         "The state of the attachment of the volume (`attaching` \\| `detaching` \\| `attached` \\| `detached`).",
										MarkdownDescription: "The state of the attachment of the volume (`attaching` \\| `detaching` \\| `attached` \\| `detached`).",
									},
									"vm_id": schema.StringAttribute{
										Computed:            true,
										Description:         "The ID of the VM.",
										MarkdownDescription: "The ID of the VM.",
									},
								},
								CustomType: LinkedVolumesType{
									ObjectType: types.ObjectType{
										AttrTypes: LinkedVolumesValue{}.AttributeTypes(ctx),
									},
								},
							},
							Computed:            true,
							Description:         "Information about your volume attachment.",
							MarkdownDescription: "Information about your volume attachment.",
						},
						"size": schema.Int64Attribute{
							Computed:            true,
							Description:         "The size of the volume, in gibibytes (GiB).",
							MarkdownDescription: "The size of the volume, in gibibytes (GiB).",
						},
						"snapshot_id": schema.StringAttribute{
							Computed:            true,
							Description:         "The snapshot from which the volume was created.",
							MarkdownDescription: "The snapshot from which the volume was created.",
						},
						"state": schema.StringAttribute{
							Computed:            true,
							Description:         "The state of the volume (`creating` \\| `available` \\| `in-use` \\| `updating` \\| `deleting` \\| `error`).",
							MarkdownDescription: "The state of the volume (`creating` \\| `available` \\| `in-use` \\| `updating` \\| `deleting` \\| `error`).",
						},
						"type": schema.StringAttribute{
							Computed:            true,
							Description:         "The type of the volume (`standard` \\| `gp2` \\| `io1`).",
							MarkdownDescription: "The type of the volume (`standard` \\| `gp2` \\| `io1`).",
						},
					},
				},
			},
			"creation_dates": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "The dates and times of creation of the volumes.",
				MarkdownDescription: "The dates and times of creation of the volumes, in ISO 8601 date-time format (for example, 2020-06-30T00:00:00.000Z).",
			},
			"link_volume_delete_on_vm_deletion": schema.BoolAttribute{
				Optional:            true,
				Description:         "Whether the volumes are deleted or not when terminating the VMs.",
				MarkdownDescription: "Whether the volumes are deleted or not when terminating the VMs.",
			},
			"link_volume_device_names": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "The VM device names.",
				MarkdownDescription: "The VM device names.",
			},
			"link_volume_link_dates": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "The dates and times of creation of the volumes.",
				MarkdownDescription: "The dates and times of creation of the volumes, in ISO 8601 date-time format (for example, 2020-06-30T00:00:00.000Z).",
			},
			"link_volume_link_states": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "The attachment states of the volumes (attaching | detaching | attached | detached).",
				MarkdownDescription: "The attachment states of the volumes (attaching | detaching | attached | detached).",
			},
			"link_volume_vm_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "One or more IDs of VMs.",
				MarkdownDescription: "One or more IDs of VMs.",
			},
			"snapshot_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "The snapshots from which the volumes were created.",
				MarkdownDescription: "The snapshots from which the volumes were created.",
			},
			"volume_sizes": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Optional:            true,
				Description:         "The sizes of the volumes, in gibibytes (GiB).",
				MarkdownDescription: "The sizes of the volumes, in gibibytes (GiB).",
			},
			"volume_states": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "The states of the volumes (creating | available | in-use | updating | deleting | error).",
				MarkdownDescription: "The states of the volumes (creating | available | in-use | updating | deleting | error).",
			},
			"volume_types": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "The types of the volumes (standard | gp2 | io1).",
				MarkdownDescription: "The types of the volumes (standard | gp2 | io1).",
			},
			"availability_zone_names": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "The names of the Subregions in which the volumes were created.",
				MarkdownDescription: "The names of the Subregions in which the volumes were created.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "The IDs of the volumes.",
				MarkdownDescription: "The IDs of the volumes.",
			},
		},
	}
}

type VolumeModel struct {
	AvailabilityZoneName types.String `tfsdk:"availability_zone_name"`
	CreationDate         types.String `tfsdk:"creation_date"`
	Id                   types.String `tfsdk:"id"`
	Iops                 types.Int64  `tfsdk:"iops"`
	LinkedVolumes        types.List   `tfsdk:"linked_volumes"`
	Size                 types.Int64  `tfsdk:"size"`
	SnapshotId           types.String `tfsdk:"snapshot_id"`
	State                types.String `tfsdk:"state"`
	Type                 types.String `tfsdk:"type"`
}

var _ basetypes.ObjectTypable = LinkedVolumesType{}

type LinkedVolumesType struct {
	basetypes.ObjectType
}

func (t LinkedVolumesType) Equal(o attr.Type) bool {
	other, ok := o.(LinkedVolumesType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t LinkedVolumesType) String() string {
	return "LinkedVolumesType"
}

func (t LinkedVolumesType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	deleteOnVmDeletionAttribute, ok := attributes["delete_on_vm_deletion"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`delete_on_vm_deletion is missing from object`)

		return nil, diags
	}

	deleteOnVmDeletionVal, ok := deleteOnVmDeletionAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`delete_on_vm_deletion expected to be basetypes.BoolValue, was: %T`, deleteOnVmDeletionAttribute))
	}

	deviceNameAttribute, ok := attributes["device_name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`device_name is missing from object`)

		return nil, diags
	}

	deviceNameVal, ok := deviceNameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`device_name expected to be basetypes.StringValue, was: %T`, deviceNameAttribute))
	}

	idAttribute, ok := attributes["id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`id is missing from object`)

		return nil, diags
	}

	idVal, ok := idAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`id expected to be basetypes.StringValue, was: %T`, idAttribute))
	}

	stateAttribute, ok := attributes["state"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`state is missing from object`)

		return nil, diags
	}

	stateVal, ok := stateAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`state expected to be basetypes.StringValue, was: %T`, stateAttribute))
	}

	vmIdAttribute, ok := attributes["vm_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`vm_id is missing from object`)

		return nil, diags
	}

	vmIdVal, ok := vmIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`vm_id expected to be basetypes.StringValue, was: %T`, vmIdAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return LinkedVolumesValue{
		DeleteOnVmDeletion: deleteOnVmDeletionVal,
		DeviceName:         deviceNameVal,
		Id:                 idVal,
		State:              stateVal,
		VmId:               vmIdVal,
		state:              attr.ValueStateKnown,
	}, diags
}

func NewLinkedVolumesValueNull() LinkedVolumesValue {
	return LinkedVolumesValue{
		state: attr.ValueStateNull,
	}
}

func NewLinkedVolumesValueUnknown() LinkedVolumesValue {
	return LinkedVolumesValue{
		state: attr.ValueStateUnknown,
	}
}

func NewLinkedVolumesValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (LinkedVolumesValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing LinkedVolumesValue Attribute Value",
				"While creating a LinkedVolumesValue value, a missing attribute value was detected. "+
					"A LinkedVolumesValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("LinkedVolumesValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid LinkedVolumesValue Attribute Type",
				"While creating a LinkedVolumesValue value, an invalid attribute value was detected. "+
					"A LinkedVolumesValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("LinkedVolumesValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("LinkedVolumesValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra LinkedVolumesValue Attribute Value",
				"While creating a LinkedVolumesValue value, an extra attribute value was detected. "+
					"A LinkedVolumesValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra LinkedVolumesValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewLinkedVolumesValueUnknown(), diags
	}

	deleteOnVmDeletionAttribute, ok := attributes["delete_on_vm_deletion"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`delete_on_vm_deletion is missing from object`)

		return NewLinkedVolumesValueUnknown(), diags
	}

	deleteOnVmDeletionVal, ok := deleteOnVmDeletionAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`delete_on_vm_deletion expected to be basetypes.BoolValue, was: %T`, deleteOnVmDeletionAttribute))
	}

	deviceNameAttribute, ok := attributes["device_name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`device_name is missing from object`)

		return NewLinkedVolumesValueUnknown(), diags
	}

	deviceNameVal, ok := deviceNameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`device_name expected to be basetypes.StringValue, was: %T`, deviceNameAttribute))
	}

	idAttribute, ok := attributes["id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`id is missing from object`)

		return NewLinkedVolumesValueUnknown(), diags
	}

	idVal, ok := idAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`id expected to be basetypes.StringValue, was: %T`, idAttribute))
	}

	stateAttribute, ok := attributes["state"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`state is missing from object`)

		return NewLinkedVolumesValueUnknown(), diags
	}

	stateVal, ok := stateAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`state expected to be basetypes.StringValue, was: %T`, stateAttribute))
	}

	vmIdAttribute, ok := attributes["vm_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`vm_id is missing from object`)

		return NewLinkedVolumesValueUnknown(), diags
	}

	vmIdVal, ok := vmIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`vm_id expected to be basetypes.StringValue, was: %T`, vmIdAttribute))
	}

	if diags.HasError() {
		return NewLinkedVolumesValueUnknown(), diags
	}

	return LinkedVolumesValue{
		DeleteOnVmDeletion: deleteOnVmDeletionVal,
		DeviceName:         deviceNameVal,
		Id:                 idVal,
		State:              stateVal,
		VmId:               vmIdVal,
		state:              attr.ValueStateKnown,
	}, diags
}

func NewLinkedVolumesValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) LinkedVolumesValue {
	object, diags := NewLinkedVolumesValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewLinkedVolumesValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t LinkedVolumesType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewLinkedVolumesValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewLinkedVolumesValueUnknown(), nil
	}

	if in.IsNull() {
		return NewLinkedVolumesValueNull(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := in.As(&val)

	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return NewLinkedVolumesValueMust(LinkedVolumesValue{}.AttributeTypes(ctx), attributes), nil
}

func (t LinkedVolumesType) ValueType(ctx context.Context) attr.Value {
	return LinkedVolumesValue{}
}

var _ basetypes.ObjectValuable = LinkedVolumesValue{}

type LinkedVolumesValue struct {
	DeleteOnVmDeletion basetypes.BoolValue   `tfsdk:"delete_on_vm_deletion"`
	DeviceName         basetypes.StringValue `tfsdk:"device_name"`
	Id                 basetypes.StringValue `tfsdk:"id"`
	State              basetypes.StringValue `tfsdk:"state"`
	VmId               basetypes.StringValue `tfsdk:"vm_id"`
	state              attr.ValueState
}

func (v LinkedVolumesValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 5)

	var val tftypes.Value
	var err error

	attrTypes["delete_on_vm_deletion"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["device_name"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["id"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["state"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["vm_id"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 5)

		val, err = v.DeleteOnVmDeletion.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["delete_on_vm_deletion"] = val

		val, err = v.DeviceName.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["device_name"] = val

		val, err = v.Id.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["id"] = val

		val, err = v.State.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["state"] = val

		val, err = v.VmId.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["vm_id"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v LinkedVolumesValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v LinkedVolumesValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v LinkedVolumesValue) String() string {
	return "LinkedVolumesValue"
}

func (v LinkedVolumesValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	objVal, diags := types.ObjectValue(
		map[string]attr.Type{
			"delete_on_vm_deletion": basetypes.BoolType{},
			"device_name":           basetypes.StringType{},
			"id":                    basetypes.StringType{},
			"state":                 basetypes.StringType{},
			"vm_id":                 basetypes.StringType{},
		},
		map[string]attr.Value{
			"delete_on_vm_deletion": v.DeleteOnVmDeletion,
			"device_name":           v.DeviceName,
			"id":                    v.Id,
			"state":                 v.State,
			"vm_id":                 v.VmId,
		})

	return objVal, diags
}

func (v LinkedVolumesValue) Equal(o attr.Value) bool {
	other, ok := o.(LinkedVolumesValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.DeleteOnVmDeletion.Equal(other.DeleteOnVmDeletion) {
		return false
	}

	if !v.DeviceName.Equal(other.DeviceName) {
		return false
	}

	if !v.Id.Equal(other.Id) {
		return false
	}

	if !v.State.Equal(other.State) {
		return false
	}

	if !v.VmId.Equal(other.VmId) {
		return false
	}

	return true
}

func (v LinkedVolumesValue) Type(ctx context.Context) attr.Type {
	return LinkedVolumesType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v LinkedVolumesValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"delete_on_vm_deletion": basetypes.BoolType{},
		"device_name":           basetypes.StringType{},
		"id":                    basetypes.StringType{},
		"state":                 basetypes.StringType{},
		"vm_id":                 basetypes.StringType{},
	}
}

var _ basetypes.ObjectTypable = TagsType{}

type TagsType struct {
	basetypes.ObjectType
}

func (t TagsType) Equal(o attr.Type) bool {
	other, ok := o.(TagsType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t TagsType) String() string {
	return "TagsType"
}

func (t TagsType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	keyAttribute, ok := attributes["key"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`key is missing from object`)

		return nil, diags
	}

	keyVal, ok := keyAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`key expected to be basetypes.StringValue, was: %T`, keyAttribute))
	}

	valueAttribute, ok := attributes["value"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`value is missing from object`)

		return nil, diags
	}

	valueVal, ok := valueAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`value expected to be basetypes.StringValue, was: %T`, valueAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return TagsValue{
		Key:   keyVal,
		Value: valueVal,
		state: attr.ValueStateKnown,
	}, diags
}

func NewTagsValueNull() TagsValue {
	return TagsValue{
		state: attr.ValueStateNull,
	}
}

func NewTagsValueUnknown() TagsValue {
	return TagsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewTagsValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (TagsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing TagsValue Attribute Value",
				"While creating a TagsValue value, a missing attribute value was detected. "+
					"A TagsValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("TagsValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid TagsValue Attribute Type",
				"While creating a TagsValue value, an invalid attribute value was detected. "+
					"A TagsValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("TagsValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("TagsValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra TagsValue Attribute Value",
				"While creating a TagsValue value, an extra attribute value was detected. "+
					"A TagsValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra TagsValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewTagsValueUnknown(), diags
	}

	keyAttribute, ok := attributes["key"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`key is missing from object`)

		return NewTagsValueUnknown(), diags
	}

	keyVal, ok := keyAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`key expected to be basetypes.StringValue, was: %T`, keyAttribute))
	}

	valueAttribute, ok := attributes["value"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`value is missing from object`)

		return NewTagsValueUnknown(), diags
	}

	valueVal, ok := valueAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`value expected to be basetypes.StringValue, was: %T`, valueAttribute))
	}

	if diags.HasError() {
		return NewTagsValueUnknown(), diags
	}

	return TagsValue{
		Key:   keyVal,
		Value: valueVal,
		state: attr.ValueStateKnown,
	}, diags
}

func NewTagsValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) TagsValue {
	object, diags := NewTagsValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewTagsValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t TagsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewTagsValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewTagsValueUnknown(), nil
	}

	if in.IsNull() {
		return NewTagsValueNull(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := in.As(&val)

	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return NewTagsValueMust(TagsValue{}.AttributeTypes(ctx), attributes), nil
}

func (t TagsType) ValueType(ctx context.Context) attr.Value {
	return TagsValue{}
}

var _ basetypes.ObjectValuable = TagsValue{}

type TagsValue struct {
	Key   basetypes.StringValue `tfsdk:"key"`
	Value basetypes.StringValue `tfsdk:"value"`
	state attr.ValueState
}

func (v TagsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 2)

	var val tftypes.Value
	var err error

	attrTypes["key"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["value"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 2)

		val, err = v.Key.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["key"] = val

		val, err = v.Value.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["value"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v TagsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v TagsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v TagsValue) String() string {
	return "TagsValue"
}

func (v TagsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	objVal, diags := types.ObjectValue(
		map[string]attr.Type{
			"key":   basetypes.StringType{},
			"value": basetypes.StringType{},
		},
		map[string]attr.Value{
			"key":   v.Key,
			"value": v.Value,
		})

	return objVal, diags
}

func (v TagsValue) Equal(o attr.Value) bool {
	other, ok := o.(TagsValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Key.Equal(other.Key) {
		return false
	}

	if !v.Value.Equal(other.Value) {
		return false
	}

	return true
}

func (v TagsValue) Type(ctx context.Context) attr.Type {
	return TagsType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v TagsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"key":   basetypes.StringType{},
		"value": basetypes.StringType{},
	}
}
