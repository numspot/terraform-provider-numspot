// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package snapshot

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
)

func SnapshotResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"is_public": schema.BoolAttribute{
						Computed:            true,
						Description:         "A global permission for all accounts.<br />\n(Request) Set this parameter to true to make the resource public (if the parent parameter is `Additions`) or to make the resource private (if the parent parameter is `Removals`).<br />\n(Response) If true, the resource is public. If false, the resource is private.",
						MarkdownDescription: "A global permission for all accounts.<br />\n(Request) Set this parameter to true to make the resource public (if the parent parameter is `Additions`) or to make the resource private (if the parent parameter is `Removals`).<br />\n(Response) If true, the resource is public. If false, the resource is private.",
					},
				},
				CustomType: AccessType{
					ObjectType: types.ObjectType{
						AttrTypes: AccessValue{}.AttributeTypes(ctx),
					},
				},
				Computed:            true,
				Description:         "Permissions for the resource.",
				MarkdownDescription: "Permissions for the resource.",
			},
			"creation_date": schema.StringAttribute{
				Computed:            true,
				Description:         "The date and time of creation of the snapshot.",
				MarkdownDescription: "The date and time of creation of the snapshot.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(), // MANUALLY EDITED : Adds RequireReplace
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "A description for the snapshot.",
				MarkdownDescription: "A description for the snapshot.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(), // MANUALLY EDITED : Adds RequireReplace
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the snapshot.",
				MarkdownDescription: "The ID of the snapshot.",
			},
			"progress": schema.Int64Attribute{
				Computed:            true,
				Description:         "The progress of the snapshot, as a percentage.",
				MarkdownDescription: "The progress of the snapshot, as a percentage.",
			},
			"source_region_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "**(when copying a snapshot)** The name of the source Region, which must be the same as the Region of your account.",
				MarkdownDescription: "**(when copying a snapshot)** The name of the source Region, which must be the same as the Region of your account.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(), // MANUALLY EDITED : Adds RequireReplace
				},
			},
			"source_snapshot_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "**(when copying a snapshot)** The ID of the snapshot you want to copy.",
				MarkdownDescription: "**(when copying a snapshot)** The ID of the snapshot you want to copy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(), // MANUALLY EDITED : Adds RequireReplace
				},
			},
			"state": schema.StringAttribute{
				Computed:            true,
				Description:         "The state of the snapshot (`in-queue` \\| `completed` \\| `error`).",
				MarkdownDescription: "The state of the snapshot (`in-queue` \\| `completed` \\| `error`).",
			},
			"tags": tags.TagsSchema(ctx), // MANUALLY EDITED : Use shared tags
			"volume_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "**(when creating from a volume)** The ID of the volume you want to create a snapshot of.",
				MarkdownDescription: "**(when creating from a volume)** The ID of the volume you want to create a snapshot of.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(), // MANUALLY EDITED : Adds RequireReplace
				},
			},
			"volume_size": schema.Int64Attribute{
				Computed:            true,
				Description:         "The size of the volume used to create the snapshot, in gibibytes (GiB).",
				MarkdownDescription: "The size of the volume used to create the snapshot, in gibibytes (GiB).",
			},
		},
		DeprecationMessage: "Managing IAAS services with Terraform is deprecated", // MANUALLY EDITED : Add Deprecation message
	}
}

type SnapshotModel struct {
	Access           AccessValue  `tfsdk:"access"`
	CreationDate     types.String `tfsdk:"creation_date"`
	Description      types.String `tfsdk:"description"`
	Id               types.String `tfsdk:"id"`
	Progress         types.Int64  `tfsdk:"progress"`
	SourceRegionName types.String `tfsdk:"source_region_name"`
	SourceSnapshotId types.String `tfsdk:"source_snapshot_id"`
	State            types.String `tfsdk:"state"`
	Tags             types.List   `tfsdk:"tags"`
	VolumeId         types.String `tfsdk:"volume_id"`
	VolumeSize       types.Int64  `tfsdk:"volume_size"`
}

var _ basetypes.ObjectTypable = AccessType{}

type AccessType struct {
	basetypes.ObjectType
}

func (t AccessType) Equal(o attr.Type) bool {
	other, ok := o.(AccessType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t AccessType) String() string {
	return "AccessType"
}

func (t AccessType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	isPublicAttribute, ok := attributes["is_public"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`is_public is missing from object`)

		return nil, diags
	}

	isPublicVal, ok := isPublicAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`is_public expected to be basetypes.BoolValue, was: %T`, isPublicAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return AccessValue{
		IsPublic: isPublicVal,
		state:    attr.ValueStateKnown,
	}, diags
}

func NewAccessValueNull() AccessValue {
	return AccessValue{
		state: attr.ValueStateNull,
	}
}

func NewAccessValueUnknown() AccessValue {
	return AccessValue{
		state: attr.ValueStateUnknown,
	}
}

func NewAccessValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (AccessValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing AccessValue Attribute Value",
				"While creating a AccessValue value, a missing attribute value was detected. "+
					"A AccessValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("AccessValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid AccessValue Attribute Type",
				"While creating a AccessValue value, an invalid attribute value was detected. "+
					"A AccessValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("AccessValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("AccessValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra AccessValue Attribute Value",
				"While creating a AccessValue value, an extra attribute value was detected. "+
					"A AccessValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra AccessValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewAccessValueUnknown(), diags
	}

	isPublicAttribute, ok := attributes["is_public"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`is_public is missing from object`)

		return NewAccessValueUnknown(), diags
	}

	isPublicVal, ok := isPublicAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`is_public expected to be basetypes.BoolValue, was: %T`, isPublicAttribute))
	}

	if diags.HasError() {
		return NewAccessValueUnknown(), diags
	}

	return AccessValue{
		IsPublic: isPublicVal,
		state:    attr.ValueStateKnown,
	}, diags
}

func NewAccessValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) AccessValue {
	object, diags := NewAccessValue(attributeTypes, attributes)

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

		panic("NewAccessValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t AccessType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewAccessValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewAccessValueUnknown(), nil
	}

	if in.IsNull() {
		return NewAccessValueNull(), nil
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

	return NewAccessValueMust(AccessValue{}.AttributeTypes(ctx), attributes), nil
}

func (t AccessType) ValueType(ctx context.Context) attr.Value {
	return AccessValue{}
}

var _ basetypes.ObjectValuable = AccessValue{}

type AccessValue struct {
	IsPublic basetypes.BoolValue `tfsdk:"is_public"`
	state    attr.ValueState
}

func (v AccessValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 1)

	var val tftypes.Value
	var err error

	attrTypes["is_public"] = basetypes.BoolType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 1)

		val, err = v.IsPublic.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["is_public"] = val

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

func (v AccessValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v AccessValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v AccessValue) String() string {
	return "AccessValue"
}

func (v AccessValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributeTypes := map[string]attr.Type{
		"is_public": basetypes.BoolType{},
	}

	if v.IsNull() {
		return types.ObjectNull(attributeTypes), diags
	}

	if v.IsUnknown() {
		return types.ObjectUnknown(attributeTypes), diags
	}

	objVal, diags := types.ObjectValue(
		attributeTypes,
		map[string]attr.Value{
			"is_public": v.IsPublic,
		})

	return objVal, diags
}

func (v AccessValue) Equal(o attr.Value) bool {
	other, ok := o.(AccessValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.IsPublic.Equal(other.IsPublic) {
		return false
	}

	return true
}

func (v AccessValue) Type(ctx context.Context) attr.Type {
	return AccessType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v AccessValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"is_public": basetypes.BoolType{},
	}
}
