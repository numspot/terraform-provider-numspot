// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package datasource_flexible_gpu

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

func FlexibleGpuDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"availability_zone_names": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The Subregions where the fGPUs are located.",
				MarkdownDescription: "The Subregions where the fGPUs are located.",
			},
			"delete_on_vm_deletion": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Indicates whether the fGPU is deleted when terminating the VM.",
				MarkdownDescription: "Indicates whether the fGPU is deleted when terminating the VM.",
			},
			"generations": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The processor generations that the fGPUs are compatible with.",
				MarkdownDescription: "The processor generations that the fGPUs are compatible with.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "One or more IDs of fGPUs.",
				MarkdownDescription: "One or more IDs of fGPUs.",
			},
			"items": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"availability_zone_name": schema.StringAttribute{
							Computed:            true,
							Description:         "The Subregion where the fGPU is located.",
							MarkdownDescription: "The Subregion where the fGPU is located.",
						},
						"delete_on_vm_deletion": schema.BoolAttribute{
							Computed:            true,
							Description:         "If true, the fGPU is deleted when the VM is terminated.",
							MarkdownDescription: "If true, the fGPU is deleted when the VM is terminated.",
						},
						"generation": schema.StringAttribute{
							Computed:            true,
							Description:         "The compatible processor generation.",
							MarkdownDescription: "The compatible processor generation.",
						},
						"id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the fGPU.",
							MarkdownDescription: "The ID of the fGPU.",
						},
						"model_name": schema.StringAttribute{
							Computed:            true,
							Description:         "The model of fGPU. For more information, see [About Flexible GPUs](https://docs.outscale.com/en/userguide/About-Flexible-GPUs.html).",
							MarkdownDescription: "The model of fGPU. For more information, see [About Flexible GPUs](https://docs.outscale.com/en/userguide/About-Flexible-GPUs.html).",
						},
						"state": schema.StringAttribute{
							Computed:            true,
							Description:         "The state of the fGPU (`allocated` \\| `attaching` \\| `attached` \\| `detaching`).",
							MarkdownDescription: "The state of the fGPU (`allocated` \\| `attaching` \\| `attached` \\| `detaching`).",
						},
						"vm_id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the VM the fGPU is attached to, if any.",
							MarkdownDescription: "The ID of the VM the fGPU is attached to, if any.",
						},
					},
					CustomType: ItemsType{
						ObjectType: types.ObjectType{
							AttrTypes: ItemsValue{}.AttributeTypes(ctx),
						},
					},
				},
				Computed:            true,
				Description:         "Information about one or more fGPUs.",
				MarkdownDescription: "Information about one or more fGPUs.",
			},
			"model_names": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "One or more models of fGPUs. For more information, see [About Flexible GPUs](https://docs.outscale.com/en/userguide/About-Flexible-GPUs.html).",
				MarkdownDescription: "One or more models of fGPUs. For more information, see [About Flexible GPUs](https://docs.outscale.com/en/userguide/About-Flexible-GPUs.html).",
			},
			"states": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The states of the fGPUs (`allocated` \\| `attaching` \\| `attached` \\| `detaching`).",
				MarkdownDescription: "The states of the fGPUs (`allocated` \\| `attaching` \\| `attached` \\| `detaching`).",
			},
			"vm_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "One or more IDs of VMs.",
				MarkdownDescription: "One or more IDs of VMs.",
			},
		},
	}
}

type FlexibleGpuModel struct {
	AvailabilityZoneName types.String `tfsdk:"availability_zone_name"`
	DeleteOnVmDeletion   types.Bool   `tfsdk:"delete_on_vm_deletion"`
	Generation           types.String `tfsdk:"generation"`
	Id                   types.String `tfsdk:"id"`
	ModelName            types.String `tfsdk:"model_name"`
	State                types.String `tfsdk:"state"`
	VmId                 types.String `tfsdk:"vm_id"`
}

var _ basetypes.ObjectTypable = ItemsType{}

type ItemsType struct {
	basetypes.ObjectType
}

func (t ItemsType) Equal(o attr.Type) bool {
	other, ok := o.(ItemsType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t ItemsType) String() string {
	return "ItemsType"
}

func (t ItemsType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	availabilityZoneNameAttribute, ok := attributes["availability_zone_name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`availability_zone_name is missing from object`)

		return nil, diags
	}

	availabilityZoneNameVal, ok := availabilityZoneNameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`availability_zone_name expected to be basetypes.StringValue, was: %T`, availabilityZoneNameAttribute))
	}

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

	generationAttribute, ok := attributes["generation"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`generation is missing from object`)

		return nil, diags
	}

	generationVal, ok := generationAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`generation expected to be basetypes.StringValue, was: %T`, generationAttribute))
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

	modelNameAttribute, ok := attributes["model_name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`model_name is missing from object`)

		return nil, diags
	}

	modelNameVal, ok := modelNameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`model_name expected to be basetypes.StringValue, was: %T`, modelNameAttribute))
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

	return ItemsValue{
		AvailabilityZoneName: availabilityZoneNameVal,
		DeleteOnVmDeletion:   deleteOnVmDeletionVal,
		Generation:           generationVal,
		Id:                   idVal,
		ModelName:            modelNameVal,
		State:                stateVal,
		VmId:                 vmIdVal,
		state:                attr.ValueStateKnown,
	}, diags
}

func NewItemsValueNull() ItemsValue {
	return ItemsValue{
		state: attr.ValueStateNull,
	}
}

func NewItemsValueUnknown() ItemsValue {
	return ItemsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewItemsValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (ItemsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing ItemsValue Attribute Value",
				"While creating a ItemsValue value, a missing attribute value was detected. "+
					"A ItemsValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("ItemsValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid ItemsValue Attribute Type",
				"While creating a ItemsValue value, an invalid attribute value was detected. "+
					"A ItemsValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("ItemsValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("ItemsValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra ItemsValue Attribute Value",
				"While creating a ItemsValue value, an extra attribute value was detected. "+
					"A ItemsValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra ItemsValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewItemsValueUnknown(), diags
	}

	availabilityZoneNameAttribute, ok := attributes["availability_zone_name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`availability_zone_name is missing from object`)

		return NewItemsValueUnknown(), diags
	}

	availabilityZoneNameVal, ok := availabilityZoneNameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`availability_zone_name expected to be basetypes.StringValue, was: %T`, availabilityZoneNameAttribute))
	}

	deleteOnVmDeletionAttribute, ok := attributes["delete_on_vm_deletion"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`delete_on_vm_deletion is missing from object`)

		return NewItemsValueUnknown(), diags
	}

	deleteOnVmDeletionVal, ok := deleteOnVmDeletionAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`delete_on_vm_deletion expected to be basetypes.BoolValue, was: %T`, deleteOnVmDeletionAttribute))
	}

	generationAttribute, ok := attributes["generation"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`generation is missing from object`)

		return NewItemsValueUnknown(), diags
	}

	generationVal, ok := generationAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`generation expected to be basetypes.StringValue, was: %T`, generationAttribute))
	}

	idAttribute, ok := attributes["id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`id is missing from object`)

		return NewItemsValueUnknown(), diags
	}

	idVal, ok := idAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`id expected to be basetypes.StringValue, was: %T`, idAttribute))
	}

	modelNameAttribute, ok := attributes["model_name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`model_name is missing from object`)

		return NewItemsValueUnknown(), diags
	}

	modelNameVal, ok := modelNameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`model_name expected to be basetypes.StringValue, was: %T`, modelNameAttribute))
	}

	stateAttribute, ok := attributes["state"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`state is missing from object`)

		return NewItemsValueUnknown(), diags
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

		return NewItemsValueUnknown(), diags
	}

	vmIdVal, ok := vmIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`vm_id expected to be basetypes.StringValue, was: %T`, vmIdAttribute))
	}

	if diags.HasError() {
		return NewItemsValueUnknown(), diags
	}

	return ItemsValue{
		AvailabilityZoneName: availabilityZoneNameVal,
		DeleteOnVmDeletion:   deleteOnVmDeletionVal,
		Generation:           generationVal,
		Id:                   idVal,
		ModelName:            modelNameVal,
		State:                stateVal,
		VmId:                 vmIdVal,
		state:                attr.ValueStateKnown,
	}, diags
}

func NewItemsValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) ItemsValue {
	object, diags := NewItemsValue(attributeTypes, attributes)

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

		panic("NewItemsValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t ItemsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewItemsValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewItemsValueUnknown(), nil
	}

	if in.IsNull() {
		return NewItemsValueNull(), nil
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

	return NewItemsValueMust(ItemsValue{}.AttributeTypes(ctx), attributes), nil
}

func (t ItemsType) ValueType(ctx context.Context) attr.Value {
	return ItemsValue{}
}

var _ basetypes.ObjectValuable = ItemsValue{}

type ItemsValue struct {
	AvailabilityZoneName basetypes.StringValue `tfsdk:"availability_zone_name"`
	DeleteOnVmDeletion   basetypes.BoolValue   `tfsdk:"delete_on_vm_deletion"`
	Generation           basetypes.StringValue `tfsdk:"generation"`
	Id                   basetypes.StringValue `tfsdk:"id"`
	ModelName            basetypes.StringValue `tfsdk:"model_name"`
	State                basetypes.StringValue `tfsdk:"state"`
	VmId                 basetypes.StringValue `tfsdk:"vm_id"`
	state                attr.ValueState
}

func (v ItemsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 7)

	var val tftypes.Value
	var err error

	attrTypes["availability_zone_name"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["delete_on_vm_deletion"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["generation"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["id"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["model_name"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["state"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["vm_id"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 7)

		val, err = v.AvailabilityZoneName.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["availability_zone_name"] = val

		val, err = v.DeleteOnVmDeletion.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["delete_on_vm_deletion"] = val

		val, err = v.Generation.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["generation"] = val

		val, err = v.Id.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["id"] = val

		val, err = v.ModelName.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["model_name"] = val

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

func (v ItemsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ItemsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ItemsValue) String() string {
	return "ItemsValue"
}

func (v ItemsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributeTypes := map[string]attr.Type{
		"availability_zone_name": basetypes.StringType{},
		"delete_on_vm_deletion":  basetypes.BoolType{},
		"generation":             basetypes.StringType{},
		"id":                     basetypes.StringType{},
		"model_name":             basetypes.StringType{},
		"state":                  basetypes.StringType{},
		"vm_id":                  basetypes.StringType{},
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
			"availability_zone_name": v.AvailabilityZoneName,
			"delete_on_vm_deletion":  v.DeleteOnVmDeletion,
			"generation":             v.Generation,
			"id":                     v.Id,
			"model_name":             v.ModelName,
			"state":                  v.State,
			"vm_id":                  v.VmId,
		})

	return objVal, diags
}

func (v ItemsValue) Equal(o attr.Value) bool {
	other, ok := o.(ItemsValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.AvailabilityZoneName.Equal(other.AvailabilityZoneName) {
		return false
	}

	if !v.DeleteOnVmDeletion.Equal(other.DeleteOnVmDeletion) {
		return false
	}

	if !v.Generation.Equal(other.Generation) {
		return false
	}

	if !v.Id.Equal(other.Id) {
		return false
	}

	if !v.ModelName.Equal(other.ModelName) {
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

func (v ItemsValue) Type(ctx context.Context) attr.Type {
	return ItemsType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v ItemsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"availability_zone_name": basetypes.StringType{},
		"delete_on_vm_deletion":  basetypes.BoolType{},
		"generation":             basetypes.StringType{},
		"id":                     basetypes.StringType{},
		"model_name":             basetypes.StringType{},
		"state":                  basetypes.StringType{},
		"vm_id":                  basetypes.StringType{},
	}
}
