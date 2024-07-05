// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_cluster

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func ClusterResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cidr": schema.StringAttribute{
				Required:            true,
				Description:         "IP addresses in CIDR notation",
				MarkdownDescription: "IP addresses in CIDR notation",
			},
			"cluster_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^[a-zA-Z0-9_]{3,64}$"), ""),
				},
			},
			"node_pools": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"gpu": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Description:         "GPU values",
							MarkdownDescription: "GPU values",
							Validators: []validator.String{
								stringvalidator.OneOf(
									"P6",
									"P100",
									"V100",
									"A100-80",
								),
							},
						},
						"name": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexp.MustCompile("^[a-zA-Z0-9_]{3,64}$"), ""),
							},
						},
						"node_count": schema.Int64Attribute{
							Required: true,
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
						"node_profile": schema.StringAttribute{
							Required:            true,
							Description:         "Node profiles",
							MarkdownDescription: "Node profiles",
							Validators: []validator.String{
								stringvalidator.OneOf(
									"SMALL",
									"MEDIUM",
									"LARGE",
									"VERY_LARGE",
								),
							},
						},
						"tina": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
					},
					CustomType: NodePoolsType{
						ObjectType: types.ObjectType{
							AttrTypes: NodePoolsValue{}.AttributeTypes(ctx),
						},
					},
				},
				Required: true,
			},
			"space_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"urls": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"api": schema.StringAttribute{
						Computed: true,
					},
					"console": schema.StringAttribute{
						Computed: true,
					},
				},
				CustomType: UrlsType{
					ObjectType: types.ObjectType{
						AttrTypes: UrlsValue{}.AttributeTypes(ctx),
					},
				},
				Computed: true,
			},
			"version": schema.StringAttribute{
				Required: true,
			},
		},
		DeprecationMessage: "Managing Openshift clusters with Terraform still not supported",
	}
}

type ClusterModel struct {
	Cidr        types.String `tfsdk:"cidr"`
	ClusterId   types.String `tfsdk:"cluster_id"`
	Description types.String `tfsdk:"description"`
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	NodePools   types.List   `tfsdk:"node_pools"`
	SpaceId     types.String `tfsdk:"space_id"`
	Urls        UrlsValue    `tfsdk:"urls"`
	Version     types.String `tfsdk:"version"`
}

var _ basetypes.ObjectTypable = NodePoolsType{}

type NodePoolsType struct {
	basetypes.ObjectType
}

func (t NodePoolsType) Equal(o attr.Type) bool {
	other, ok := o.(NodePoolsType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t NodePoolsType) String() string {
	return "NodePoolsType"
}

func (t NodePoolsType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	gpuAttribute, ok := attributes["gpu"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`gpu is missing from object`)

		return nil, diags
	}

	gpuVal, ok := gpuAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`gpu expected to be basetypes.StringValue, was: %T`, gpuAttribute))
	}

	nameAttribute, ok := attributes["name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`name is missing from object`)

		return nil, diags
	}

	nameVal, ok := nameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`name expected to be basetypes.StringValue, was: %T`, nameAttribute))
	}

	nodeCountAttribute, ok := attributes["node_count"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`node_count is missing from object`)

		return nil, diags
	}

	nodeCountVal, ok := nodeCountAttribute.(basetypes.Int64Value)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`node_count expected to be basetypes.Int64Value, was: %T`, nodeCountAttribute))
	}

	nodeProfileAttribute, ok := attributes["node_profile"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`node_profile is missing from object`)

		return nil, diags
	}

	nodeProfileVal, ok := nodeProfileAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`node_profile expected to be basetypes.StringValue, was: %T`, nodeProfileAttribute))
	}

	tinaAttribute, ok := attributes["tina"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`tina is missing from object`)

		return nil, diags
	}

	tinaVal, ok := tinaAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`tina expected to be basetypes.StringValue, was: %T`, tinaAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return NodePoolsValue{
		Gpu:         gpuVal,
		Name:        nameVal,
		NodeCount:   nodeCountVal,
		NodeProfile: nodeProfileVal,
		Tina:        tinaVal,
		state:       attr.ValueStateKnown,
	}, diags
}

func NewNodePoolsValueNull() NodePoolsValue {
	return NodePoolsValue{
		state: attr.ValueStateNull,
	}
}

func NewNodePoolsValueUnknown() NodePoolsValue {
	return NodePoolsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewNodePoolsValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (NodePoolsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing NodePoolsValue Attribute Value",
				"While creating a NodePoolsValue value, a missing attribute value was detected. "+
					"A NodePoolsValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("NodePoolsValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid NodePoolsValue Attribute Type",
				"While creating a NodePoolsValue value, an invalid attribute value was detected. "+
					"A NodePoolsValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("NodePoolsValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("NodePoolsValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra NodePoolsValue Attribute Value",
				"While creating a NodePoolsValue value, an extra attribute value was detected. "+
					"A NodePoolsValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra NodePoolsValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewNodePoolsValueUnknown(), diags
	}

	gpuAttribute, ok := attributes["gpu"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`gpu is missing from object`)

		return NewNodePoolsValueUnknown(), diags
	}

	gpuVal, ok := gpuAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`gpu expected to be basetypes.StringValue, was: %T`, gpuAttribute))
	}

	nameAttribute, ok := attributes["name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`name is missing from object`)

		return NewNodePoolsValueUnknown(), diags
	}

	nameVal, ok := nameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`name expected to be basetypes.StringValue, was: %T`, nameAttribute))
	}

	nodeCountAttribute, ok := attributes["node_count"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`node_count is missing from object`)

		return NewNodePoolsValueUnknown(), diags
	}

	nodeCountVal, ok := nodeCountAttribute.(basetypes.Int64Value)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`node_count expected to be basetypes.Int64Value, was: %T`, nodeCountAttribute))
	}

	nodeProfileAttribute, ok := attributes["node_profile"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`node_profile is missing from object`)

		return NewNodePoolsValueUnknown(), diags
	}

	nodeProfileVal, ok := nodeProfileAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`node_profile expected to be basetypes.StringValue, was: %T`, nodeProfileAttribute))
	}

	tinaAttribute, ok := attributes["tina"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`tina is missing from object`)

		return NewNodePoolsValueUnknown(), diags
	}

	tinaVal, ok := tinaAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`tina expected to be basetypes.StringValue, was: %T`, tinaAttribute))
	}

	if diags.HasError() {
		return NewNodePoolsValueUnknown(), diags
	}

	return NodePoolsValue{
		Gpu:         gpuVal,
		Name:        nameVal,
		NodeCount:   nodeCountVal,
		NodeProfile: nodeProfileVal,
		Tina:        tinaVal,
		state:       attr.ValueStateKnown,
	}, diags
}

func NewNodePoolsValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) NodePoolsValue {
	object, diags := NewNodePoolsValue(attributeTypes, attributes)

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

		panic("NewNodePoolsValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t NodePoolsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewNodePoolsValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewNodePoolsValueUnknown(), nil
	}

	if in.IsNull() {
		return NewNodePoolsValueNull(), nil
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

	return NewNodePoolsValueMust(NodePoolsValue{}.AttributeTypes(ctx), attributes), nil
}

func (t NodePoolsType) ValueType(ctx context.Context) attr.Value {
	return NodePoolsValue{}
}

var _ basetypes.ObjectValuable = NodePoolsValue{}

type NodePoolsValue struct {
	Gpu         basetypes.StringValue `tfsdk:"gpu"`
	Name        basetypes.StringValue `tfsdk:"name"`
	NodeCount   basetypes.Int64Value  `tfsdk:"node_count"`
	NodeProfile basetypes.StringValue `tfsdk:"node_profile"`
	Tina        basetypes.StringValue `tfsdk:"tina"`
	state       attr.ValueState
}

func (v NodePoolsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 5)

	var val tftypes.Value
	var err error

	attrTypes["gpu"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["name"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["node_count"] = basetypes.Int64Type{}.TerraformType(ctx)
	attrTypes["node_profile"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["tina"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 5)

		val, err = v.Gpu.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["gpu"] = val

		val, err = v.Name.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["name"] = val

		val, err = v.NodeCount.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["node_count"] = val

		val, err = v.NodeProfile.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["node_profile"] = val

		val, err = v.Tina.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["tina"] = val

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

func (v NodePoolsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v NodePoolsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v NodePoolsValue) String() string {
	return "NodePoolsValue"
}

func (v NodePoolsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributeTypes := map[string]attr.Type{
		"gpu":          basetypes.StringType{},
		"name":         basetypes.StringType{},
		"node_count":   basetypes.Int64Type{},
		"node_profile": basetypes.StringType{},
		"tina":         basetypes.StringType{},
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
			"gpu":          v.Gpu,
			"name":         v.Name,
			"node_count":   v.NodeCount,
			"node_profile": v.NodeProfile,
			"tina":         v.Tina,
		})

	return objVal, diags
}

func (v NodePoolsValue) Equal(o attr.Value) bool {
	other, ok := o.(NodePoolsValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Gpu.Equal(other.Gpu) {
		return false
	}

	if !v.Name.Equal(other.Name) {
		return false
	}

	if !v.NodeCount.Equal(other.NodeCount) {
		return false
	}

	if !v.NodeProfile.Equal(other.NodeProfile) {
		return false
	}

	if !v.Tina.Equal(other.Tina) {
		return false
	}

	return true
}

func (v NodePoolsValue) Type(ctx context.Context) attr.Type {
	return NodePoolsType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v NodePoolsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"gpu":          basetypes.StringType{},
		"name":         basetypes.StringType{},
		"node_count":   basetypes.Int64Type{},
		"node_profile": basetypes.StringType{},
		"tina":         basetypes.StringType{},
	}
}

var _ basetypes.ObjectTypable = UrlsType{}

type UrlsType struct {
	basetypes.ObjectType
}

func (t UrlsType) Equal(o attr.Type) bool {
	other, ok := o.(UrlsType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t UrlsType) String() string {
	return "UrlsType"
}

func (t UrlsType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	apiAttribute, ok := attributes["api"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`api is missing from object`)

		return nil, diags
	}

	apiVal, ok := apiAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`api expected to be basetypes.StringValue, was: %T`, apiAttribute))
	}

	consoleAttribute, ok := attributes["console"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`console is missing from object`)

		return nil, diags
	}

	consoleVal, ok := consoleAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`console expected to be basetypes.StringValue, was: %T`, consoleAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return UrlsValue{
		Api:     apiVal,
		Console: consoleVal,
		state:   attr.ValueStateKnown,
	}, diags
}

func NewUrlsValueNull() UrlsValue {
	return UrlsValue{
		state: attr.ValueStateNull,
	}
}

func NewUrlsValueUnknown() UrlsValue {
	return UrlsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewUrlsValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (UrlsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing UrlsValue Attribute Value",
				"While creating a UrlsValue value, a missing attribute value was detected. "+
					"A UrlsValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("UrlsValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid UrlsValue Attribute Type",
				"While creating a UrlsValue value, an invalid attribute value was detected. "+
					"A UrlsValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("UrlsValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("UrlsValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra UrlsValue Attribute Value",
				"While creating a UrlsValue value, an extra attribute value was detected. "+
					"A UrlsValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra UrlsValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewUrlsValueUnknown(), diags
	}

	apiAttribute, ok := attributes["api"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`api is missing from object`)

		return NewUrlsValueUnknown(), diags
	}

	apiVal, ok := apiAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`api expected to be basetypes.StringValue, was: %T`, apiAttribute))
	}

	consoleAttribute, ok := attributes["console"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`console is missing from object`)

		return NewUrlsValueUnknown(), diags
	}

	consoleVal, ok := consoleAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`console expected to be basetypes.StringValue, was: %T`, consoleAttribute))
	}

	if diags.HasError() {
		return NewUrlsValueUnknown(), diags
	}

	return UrlsValue{
		Api:     apiVal,
		Console: consoleVal,
		state:   attr.ValueStateKnown,
	}, diags
}

func NewUrlsValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) UrlsValue {
	object, diags := NewUrlsValue(attributeTypes, attributes)

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

		panic("NewUrlsValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t UrlsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewUrlsValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewUrlsValueUnknown(), nil
	}

	if in.IsNull() {
		return NewUrlsValueNull(), nil
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

	return NewUrlsValueMust(UrlsValue{}.AttributeTypes(ctx), attributes), nil
}

func (t UrlsType) ValueType(ctx context.Context) attr.Value {
	return UrlsValue{}
}

var _ basetypes.ObjectValuable = UrlsValue{}

type UrlsValue struct {
	Api     basetypes.StringValue `tfsdk:"api"`
	Console basetypes.StringValue `tfsdk:"console"`
	state   attr.ValueState
}

func (v UrlsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 2)

	var val tftypes.Value
	var err error

	attrTypes["api"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["console"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 2)

		val, err = v.Api.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["api"] = val

		val, err = v.Console.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["console"] = val

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

func (v UrlsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v UrlsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v UrlsValue) String() string {
	return "UrlsValue"
}

func (v UrlsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributeTypes := map[string]attr.Type{
		"api":     basetypes.StringType{},
		"console": basetypes.StringType{},
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
			"api":     v.Api,
			"console": v.Console,
		})

	return objVal, diags
}

func (v UrlsValue) Equal(o attr.Value) bool {
	other, ok := o.(UrlsValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Api.Equal(other.Api) {
		return false
	}

	if !v.Console.Equal(other.Console) {
		return false
	}

	return true
}

func (v UrlsValue) Type(ctx context.Context) attr.Type {
	return UrlsType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v UrlsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"api":     basetypes.StringType{},
		"console": basetypes.StringType{},
	}
}
