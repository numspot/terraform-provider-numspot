// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_snapshot

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func SnapshotResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"account_alias": schema.StringAttribute{
				Computed:            true,
				Description:         "The account alias of the owner of the snapshot.",
				MarkdownDescription: "The account alias of the owner of the snapshot.",
			},
			"account_id": schema.StringAttribute{
				Computed:            true,
				Description:         "The account ID of the owner of the snapshot.",
				MarkdownDescription: "The account ID of the owner of the snapshot.",
			},
			"creation_date": schema.StringAttribute{
				Computed:            true,
				Description:         "The date and time of creation of the snapshot.",
				MarkdownDescription: "The date and time of creation of the snapshot.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "A description for the snapshot.",
				MarkdownDescription: "A description for the snapshot.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the snapshot.",
				MarkdownDescription: "The ID of the snapshot.",
			},
			"permissions_to_create_volume": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"account_ids": schema.ListAttribute{
						ElementType:         types.StringType,
						Computed:            true,
						Description:         "One or more account IDs that the permission is associated with.",
						MarkdownDescription: "One or more account IDs that the permission is associated with.",
					},
					"global_permission": schema.BoolAttribute{
						Computed:            true,
						Description:         "A global permission for all accounts.<br />\n(Request) Set this parameter to true to make the resource public (if the parent parameter is `Additions`) or to make the resource private (if the parent parameter is `Removals`).<br />\n(Response) If true, the resource is public. If false, the resource is private.",
						MarkdownDescription: "A global permission for all accounts.<br />\n(Request) Set this parameter to true to make the resource public (if the parent parameter is `Additions`) or to make the resource private (if the parent parameter is `Removals`).<br />\n(Response) If true, the resource is public. If false, the resource is private.",
					},
				},
				CustomType: PermissionsToCreateVolumeType{
					ObjectType: types.ObjectType{
						AttrTypes: PermissionsToCreateVolumeValue{}.AttributeTypes(ctx),
					},
				},
				Computed:            true,
				Description:         "Permissions for the resource.",
				MarkdownDescription: "Permissions for the resource.",
			},
			"progress": schema.Int64Attribute{
				Computed:            true,
				Description:         "The progress of the snapshot, as a percentage.",
				MarkdownDescription: "The progress of the snapshot, as a percentage.",
			},
			"source_region_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "(When copying) The name of the source Region, which must be the same as the Region of your account.",
				MarkdownDescription: "(When copying) The name of the source Region, which must be the same as the Region of your account.",
			},
			"source_snapshot_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "(When copying) The ID of the snapshot you want to copy.",
				MarkdownDescription: "(When copying) The ID of the snapshot you want to copy.",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				Description:         "The state of the snapshot (`in-queue` \\| `completed` \\| `error`).",
				MarkdownDescription: "The state of the snapshot (`in-queue` \\| `completed` \\| `error`).",
			},
			"volume_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "(When creating) The ID of the volume you want to create a snapshot of.",
				MarkdownDescription: "(When creating) The ID of the volume you want to create a snapshot of.",
			},
			"volume_size": schema.Int64Attribute{
				Computed:            true,
				Description:         "The size of the volume used to create the snapshot, in gibibytes (GiB).",
				MarkdownDescription: "The size of the volume used to create the snapshot, in gibibytes (GiB).",
			},
		},
	}
}

type SnapshotModel struct {
	AccountAlias              types.String                   `tfsdk:"account_alias"`
	AccountId                 types.String                   `tfsdk:"account_id"`
	CreationDate              types.String                   `tfsdk:"creation_date"`
	Description               types.String                   `tfsdk:"description"`
	Id                        types.String                   `tfsdk:"id"`
	PermissionsToCreateVolume PermissionsToCreateVolumeValue `tfsdk:"permissions_to_create_volume"`
	Progress                  types.Int64                    `tfsdk:"progress"`
	SourceRegionName          types.String                   `tfsdk:"source_region_name"`
	SourceSnapshotId          types.String                   `tfsdk:"source_snapshot_id"`
	State                     types.String                   `tfsdk:"state"`
	VolumeId                  types.String                   `tfsdk:"volume_id"`
	VolumeSize                types.Int64                    `tfsdk:"volume_size"`
}

var _ basetypes.ObjectTypable = PermissionsToCreateVolumeType{}

type PermissionsToCreateVolumeType struct {
	basetypes.ObjectType
}

func (t PermissionsToCreateVolumeType) Equal(o attr.Type) bool {
	other, ok := o.(PermissionsToCreateVolumeType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t PermissionsToCreateVolumeType) String() string {
	return "PermissionsToCreateVolumeType"
}

func (t PermissionsToCreateVolumeType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	accountIdsAttribute, ok := attributes["account_ids"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`account_ids is missing from object`)

		return nil, diags
	}

	accountIdsVal, ok := accountIdsAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`account_ids expected to be basetypes.ListValue, was: %T`, accountIdsAttribute))
	}

	globalPermissionAttribute, ok := attributes["global_permission"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`global_permission is missing from object`)

		return nil, diags
	}

	globalPermissionVal, ok := globalPermissionAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`global_permission expected to be basetypes.BoolValue, was: %T`, globalPermissionAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return PermissionsToCreateVolumeValue{
		AccountIds:       accountIdsVal,
		GlobalPermission: globalPermissionVal,
		state:            attr.ValueStateKnown,
	}, diags
}

func NewPermissionsToCreateVolumeValueNull() PermissionsToCreateVolumeValue {
	return PermissionsToCreateVolumeValue{
		state: attr.ValueStateNull,
	}
}

func NewPermissionsToCreateVolumeValueUnknown() PermissionsToCreateVolumeValue {
	return PermissionsToCreateVolumeValue{
		state: attr.ValueStateUnknown,
	}
}

func NewPermissionsToCreateVolumeValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (PermissionsToCreateVolumeValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing PermissionsToCreateVolumeValue Attribute Value",
				"While creating a PermissionsToCreateVolumeValue value, a missing attribute value was detected. "+
					"A PermissionsToCreateVolumeValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("PermissionsToCreateVolumeValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid PermissionsToCreateVolumeValue Attribute Type",
				"While creating a PermissionsToCreateVolumeValue value, an invalid attribute value was detected. "+
					"A PermissionsToCreateVolumeValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("PermissionsToCreateVolumeValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("PermissionsToCreateVolumeValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra PermissionsToCreateVolumeValue Attribute Value",
				"While creating a PermissionsToCreateVolumeValue value, an extra attribute value was detected. "+
					"A PermissionsToCreateVolumeValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra PermissionsToCreateVolumeValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewPermissionsToCreateVolumeValueUnknown(), diags
	}

	accountIdsAttribute, ok := attributes["account_ids"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`account_ids is missing from object`)

		return NewPermissionsToCreateVolumeValueUnknown(), diags
	}

	accountIdsVal, ok := accountIdsAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`account_ids expected to be basetypes.ListValue, was: %T`, accountIdsAttribute))
	}

	globalPermissionAttribute, ok := attributes["global_permission"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`global_permission is missing from object`)

		return NewPermissionsToCreateVolumeValueUnknown(), diags
	}

	globalPermissionVal, ok := globalPermissionAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`global_permission expected to be basetypes.BoolValue, was: %T`, globalPermissionAttribute))
	}

	if diags.HasError() {
		return NewPermissionsToCreateVolumeValueUnknown(), diags
	}

	return PermissionsToCreateVolumeValue{
		AccountIds:       accountIdsVal,
		GlobalPermission: globalPermissionVal,
		state:            attr.ValueStateKnown,
	}, diags
}

func NewPermissionsToCreateVolumeValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) PermissionsToCreateVolumeValue {
	object, diags := NewPermissionsToCreateVolumeValue(attributeTypes, attributes)

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

		panic("NewPermissionsToCreateVolumeValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t PermissionsToCreateVolumeType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewPermissionsToCreateVolumeValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewPermissionsToCreateVolumeValueUnknown(), nil
	}

	if in.IsNull() {
		return NewPermissionsToCreateVolumeValueNull(), nil
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

	return NewPermissionsToCreateVolumeValueMust(PermissionsToCreateVolumeValue{}.AttributeTypes(ctx), attributes), nil
}

func (t PermissionsToCreateVolumeType) ValueType(ctx context.Context) attr.Value {
	return PermissionsToCreateVolumeValue{}
}

var _ basetypes.ObjectValuable = PermissionsToCreateVolumeValue{}

type PermissionsToCreateVolumeValue struct {
	AccountIds       basetypes.ListValue `tfsdk:"account_ids"`
	GlobalPermission basetypes.BoolValue `tfsdk:"global_permission"`
	state            attr.ValueState
}

func (v PermissionsToCreateVolumeValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 2)

	var val tftypes.Value
	var err error

	attrTypes["account_ids"] = basetypes.ListType{
		ElemType: types.StringType,
	}.TerraformType(ctx)
	attrTypes["global_permission"] = basetypes.BoolType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 2)

		val, err = v.AccountIds.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["account_ids"] = val

		val, err = v.GlobalPermission.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["global_permission"] = val

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

func (v PermissionsToCreateVolumeValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v PermissionsToCreateVolumeValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v PermissionsToCreateVolumeValue) String() string {
	return "PermissionsToCreateVolumeValue"
}

func (v PermissionsToCreateVolumeValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	accountIdsVal, d := types.ListValue(types.StringType, v.AccountIds.Elements())

	diags.Append(d...)

	if d.HasError() {
		return types.ObjectUnknown(map[string]attr.Type{
			"account_ids": basetypes.ListType{
				ElemType: types.StringType,
			},
			"global_permission": basetypes.BoolType{},
		}), diags
	}

	objVal, diags := types.ObjectValue(
		map[string]attr.Type{
			"account_ids": basetypes.ListType{
				ElemType: types.StringType,
			},
			"global_permission": basetypes.BoolType{},
		},
		map[string]attr.Value{
			"account_ids":       accountIdsVal,
			"global_permission": v.GlobalPermission,
		})

	return objVal, diags
}

func (v PermissionsToCreateVolumeValue) Equal(o attr.Value) bool {
	other, ok := o.(PermissionsToCreateVolumeValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.AccountIds.Equal(other.AccountIds) {
		return false
	}

	if !v.GlobalPermission.Equal(other.GlobalPermission) {
		return false
	}

	return true
}

func (v PermissionsToCreateVolumeValue) Type(ctx context.Context) attr.Type {
	return PermissionsToCreateVolumeType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v PermissionsToCreateVolumeValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"account_ids": basetypes.ListType{
			ElemType: types.StringType,
		},
		"global_permission": basetypes.BoolType{},
	}
}