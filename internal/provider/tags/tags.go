package tags

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func TagsSchema(ctx context.Context) schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"key": schema.StringAttribute{
					Required:            true,
					Description:         "The key of the tag, with a minimum of 1 character.",
					MarkdownDescription: "The key of the tag, with a minimum of 1 character.",
				},
				"value": schema.StringAttribute{
					Required:            true,
					Description:         "The value of the tag, between 0 and 255 characters.",
					MarkdownDescription: "The value of the tag, between 0 and 255 characters.",
				},
			},
			CustomType: TagsType{
				ObjectType: types.ObjectType{
					AttrTypes: TagsValue{}.AttributeTypes(ctx),
				},
			},
		},
		Computed:            true,
		Optional:            true,
		Description:         "One or more tags associated with the resource.",
		MarkdownDescription: "One or more tags associated with the resource.",
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

// Diff calculates the differences between two slices of tags: which tags to create, delete, and update.
// Assumes that a tag's Key is unique in the slice.
func Diff(current, desired []TagsValue) (toCreate, toDelete, toUpdate []TagsValue) {
	currentMap := make(map[string]TagsValue)
	desiredMap := make(map[string]TagsValue)

	for _, tag := range current {
		currentMap[tag.Key.ValueString()] = tag
	}

	for _, tag := range desired {
		desiredMap[tag.Key.ValueString()] = tag
		if _, exists := currentMap[tag.Key.ValueString()]; !exists {
			toCreate = append(toCreate, tag)
		} else if currentMap[tag.Key.ValueString()].Value != tag.Value {
			toUpdate = append(toUpdate, tag)
		}
	}

	for _, tag := range current {
		if _, exists := desiredMap[tag.Key.ValueString()]; !exists {
			toDelete = append(toDelete, tag)
		}
	}

	return toCreate, toDelete, toUpdate
}

func CreateTagsFromTf(
	ctx context.Context,
	apiClient *iaas.ClientWithResponses,
	spaceId iaas.SpaceId,
	diagnostics *diag.Diagnostics,
	resourceId string,
	tags types.List,
) {
	tfTags := make([]TagsValue, 0, len(tags.Elements()))
	tags.ElementsAs(ctx, &tfTags, false)

	apiTags := make([]iaas.ResourceTag, 0, len(tfTags))
	for _, tfTag := range tfTags {
		apiTags = append(apiTags, iaas.ResourceTag{
			Key:   tfTag.Key.ValueString(),
			Value: tfTag.Value.ValueString(),
		})
	}

	CreateTags(ctx, apiClient, spaceId, diagnostics, resourceId, apiTags)
}

func CreateTags(
	ctx context.Context,
	apiClient *iaas.ClientWithResponses,
	spaceId iaas.SpaceId,
	diagnostics *diag.Diagnostics,
	resourceId string,
	tags []iaas.ResourceTag,
) {
	res, err := apiClient.CreateTagsWithResponse(ctx, spaceId, iaas.CreateTagsJSONRequestBody{
		ResourceIds: []string{resourceId},
		Tags:        tags,
	})
	if err != nil {
		diagnostics.AddError("Failed to create tags", err.Error())
		return
	}

	if res.StatusCode() != http.StatusNoContent {
		apiError := utils.HandleError(res.Body)
		diagnostics.AddError("Failed to create Tags", apiError.Error())
		return
	}
}

func tagFromAPI(ctx context.Context, tag iaas.Tag) (TagsValue, diag.Diagnostics) {
	return NewTagsValue(
		TagsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"key":   types.StringPointerValue(tag.Key),
			"value": types.StringPointerValue(tag.Value),
		},
	)
}

func ResourceTagFromAPI(ctx context.Context, tag iaas.ResourceTag) (TagsValue, diag.Diagnostics) {
	return NewTagsValue(
		TagsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"key":   types.StringValue(tag.Key),
			"value": types.StringValue(tag.Value),
		},
	)
}

func ReadTags(
	ctx context.Context,
	apiClient *iaas.ClientWithResponses,
	spaceId iaas.SpaceId,
	diagnostics diag.Diagnostics,
	resourceId string,
) types.List {
	resourceIds := []string{resourceId}
	res, err := apiClient.ReadTagsWithResponse(ctx, spaceId, &iaas.ReadTagsParams{
		ResourceIds: &resourceIds,
	})
	if err != nil {
		diagnostics.AddError("Failed to read Tags", err.Error())
		return types.List{}
	}

	if res.StatusCode() != http.StatusOK {
		apiError := utils.HandleError(res.Body)
		diagnostics.AddError("Failed to read Tags", apiError.Error())
		return types.List{}
	}

	if res.JSON200 == nil || res.JSON200.Items == nil {
		diagnostics.AddError("Failed to read Tags", "response body is null.")
		return types.List{}
	}

	tfTags, resDiagnostics := utils.GenericListToTfListValue(
		ctx,
		TagsValue{},
		tagFromAPI,
		*res.JSON200.Items,
	)
	if resDiagnostics.HasError() {
		diagnostics.Append(resDiagnostics...)
		return types.List{}
	}

	return tfTags
}

func DeleteTags(
	ctx context.Context,
	apiClient *iaas.ClientWithResponses,
	spaceId iaas.SpaceId,
	diagnostics *diag.Diagnostics,
	resourceId string,
	tags []iaas.ResourceTag,
) {
	res, err := apiClient.DeleteTagsWithResponse(ctx, spaceId, iaas.DeleteTagsJSONRequestBody{
		ResourceIds: []string{resourceId},
		Tags:        tags,
	})
	if err != nil {
		diagnostics.AddError("Failed to delete Tag", err.Error())
		return
	}

	if res.StatusCode() != http.StatusNoContent {
		apiError := utils.HandleError(res.Body)
		diagnostics.AddError("Failed to delete Tags", apiError.Error())
		return
	}
}

func UpdateTags(
	ctx context.Context,
	stateTagsTf, planTagsTf types.List,
	diagnostics *diag.Diagnostics,
	apiClient *iaas.ClientWithResponses,
	spaceId iaas.SpaceId,
	resourceId string,
) {
	var (
		stateTags []TagsValue
		planTags  []TagsValue
	)

	diags := stateTagsTf.ElementsAs(ctx, &stateTags, false)
	if diags.HasError() {
		diagnostics.Append(diags...)
		return
	}

	diags = planTagsTf.ElementsAs(ctx, &planTags, false)
	if diags.HasError() {
		diagnostics.Append(diags...)
		return
	}

	toCreate, toDelete, toUpdate := Diff(stateTags, planTags)

	toDeleteApiTags := make([]iaas.ResourceTag, 0, len(toUpdate)+len(toDelete))
	toCreateApiTags := make([]iaas.ResourceTag, 0, len(toUpdate)+len(toCreate))
	for _, e := range toCreate {
		toCreateApiTags = append(toCreateApiTags, iaas.ResourceTag{
			Key:   e.Key.ValueString(),
			Value: e.Value.ValueString(),
		})
	}

	for _, e := range toDelete {
		toDeleteApiTags = append(toDeleteApiTags, iaas.ResourceTag{
			Key:   e.Key.ValueString(),
			Value: e.Value.ValueString(),
		})
	}

	for _, e := range toUpdate {
		// Delete
		toDeleteApiTags = append(toDeleteApiTags, iaas.ResourceTag{
			Key:   e.Key.ValueString(),
			Value: e.Value.ValueString(),
		})

		// Create
		toCreateApiTags = append(toCreateApiTags, iaas.ResourceTag{
			Key:   e.Key.ValueString(),
			Value: e.Value.ValueString(),
		})
	}

	DeleteTags(
		ctx,
		apiClient,
		spaceId,
		diagnostics,
		resourceId,
		toDeleteApiTags,
	)

	if diagnostics.HasError() {
		return
	}

	CreateTags(
		ctx,
		apiClient,
		spaceId,
		diagnostics,
		resourceId,
		toCreateApiTags,
	)
}
