package acl

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func ACLsResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"space_id": schema.StringAttribute{
				Required:            true,
				Description:         "Space ID",
				MarkdownDescription: "Space ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_account_id": schema.StringAttribute{
				Required:            true,
				Description:         "Service account ID",
				MarkdownDescription: "Service account ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service": schema.StringAttribute{
				Required:            true,
				Description:         "Name of the service making the call",
				MarkdownDescription: "Name of the service making the call",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource": schema.StringAttribute{
				Required:            true,
				Description:         "Type of the resource being accessed",
				MarkdownDescription: "Type of the resource being accessed",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subresource": schema.StringAttribute{
				Optional:            true,
				Description:         "Specific type of the subresource within the main resource",
				MarkdownDescription: "Specific type of the subresource within the main resource",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"acls": schema.SetNestedAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "List of ACLs",
				MarkdownDescription: "List of ACLs",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"resource_id": schema.StringAttribute{
							Required:            true,
							Description:         "Unique identifier of a resource",
							MarkdownDescription: "Unique identifier of a resource",
						},
						"permission_id": schema.StringAttribute{
							Required:            true,
							Description:         "ID of the permission",
							MarkdownDescription: "ID of the permission",
						},
					},
					CustomType: ACLType{
						ObjectType: types.ObjectType{
							AttrTypes: ACLValue{}.AttributeTypes(ctx),
						},
					},
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

type ACLsModel struct {
	SpaceId          types.String `tfsdk:"space_id"`
	ServiceAccountId types.String `tfsdk:"service_account_id"`
	Service          types.String `tfsdk:"service"`
	Resource         types.String `tfsdk:"resource"`
	Subresource      types.String `tfsdk:"subresource"`
	ACLs             types.Set    `tfsdk:"acls"`
}

var _ basetypes.ObjectTypable = ACLType{}

type ACLType struct {
	basetypes.ObjectType
}

func (t ACLType) Equal(o attr.Type) bool {
	other, ok := o.(ACLType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t ACLType) String() string {
	return "ACLType"
}

func (t ACLType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	resourceIdAttribute, ok := attributes["resource_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`resourceId is missing from object`)

		return nil, diags
	}

	resourceIdVal, ok := resourceIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`resourceId expected to be basetypes.StringValue, was: %T`, resourceIdAttribute))
	}

	permissionIdAttribute, ok := attributes["permission_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`permissionId is missing from object`)

		return nil, diags
	}

	permissionIdVal, ok := permissionIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`permissionId expected to be basetypes.StringValue, was: %T`, permissionIdAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return ACLValue{
		ResourceId:   resourceIdVal,
		PermissionId: permissionIdVal,
		state:        attr.ValueStateKnown,
	}, diags
}

func (t ACLType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewACLValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewACLValueUnknown(), nil
	}

	if in.IsNull() {
		return NewACLValueNull(), nil
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

	return NewACLValueMust(ACLValue{}.AttributeTypes(ctx), attributes), nil
}

func (t ACLType) ValueType(ctx context.Context) attr.Value {
	return ACLValue{}
}

func NewACLValueNull() ACLValue {
	return ACLValue{
		state: attr.ValueStateNull,
	}
}

func NewACLValueUnknown() ACLValue {
	return ACLValue{
		state: attr.ValueStateUnknown,
	}
}

func NewACLValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (ACLValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing ACLValue Attribute Value",
				"While creating a ACLValue value, a missing attribute value was detected. "+
					"A ACLValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("ACLValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid ACLValue Attribute Type",
				"While creating a ACLValue value, an invalid attribute value was detected. "+
					"A ACLValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("ACLValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("ACLValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra ACLValue Attribute Value",
				"While creating a ACLValue value, an extra attribute value was detected. "+
					"A ACLValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra ACLValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewACLValueUnknown(), diags
	}

	resourceIdAttribute, ok := attributes["resource_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`resource_id is missing from object`)

		return NewACLValueUnknown(), diags
	}

	resourceIdVal, ok := resourceIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`id expected to be basetypes.StringValue, was: %T`, resourceIdAttribute))
	}

	permissionIdAttribute, ok := attributes["permission_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`permission_id is missing from object`)

		return NewACLValueUnknown(), diags
	}

	permissionIdVal, ok := permissionIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`permission_id expected to be basetypes.StringValue, was: %T`, permissionIdAttribute))
	}

	return ACLValue{
		ResourceId:   resourceIdVal,
		PermissionId: permissionIdVal,
		state:        attr.ValueStateKnown,
	}, diags
}

func NewACLValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) ACLValue {
	object, diags := NewACLValue(attributeTypes, attributes)

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

		panic("NewACLValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

type ACLValue struct {
	ResourceId   basetypes.StringValue `tfsdk:"resource_id"`
	PermissionId basetypes.StringValue `tfsdk:"permission_id"`
	state        attr.ValueState
}

func (v ACLValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"resource_id":   basetypes.StringType{},
		"permission_id": basetypes.StringType{},
	}
}

func (v ACLValue) Type(ctx context.Context) attr.Type {
	return ACLType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v ACLValue) Equal(o attr.Value) bool {
	other, ok := o.(ACLValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.PermissionId.Equal(other.PermissionId) {
		return false
	}

	if !v.ResourceId.Equal(other.ResourceId) {
		return false
	}

	return true
}

func (v ACLValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ACLValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ACLValue) String() string {
	return "ACLValue"
}

func (v ACLValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 2)

	var val tftypes.Value
	var err error

	attrTypes["resource_id"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["permission_id"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 2)

		val, err = v.PermissionId.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["permission_id"] = val

		val, err = v.ResourceId.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["resource_id"] = val

		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v ACLValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributeTypes := map[string]attr.Type{
		"resource_id":   basetypes.StringType{},
		"permission_id": basetypes.StringType{},
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
			"resource_id":   v.ResourceId,
			"permission_id": v.PermissionId,
		})

	return objVal, diags
}
