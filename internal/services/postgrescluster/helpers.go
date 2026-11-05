package postgrescluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services/postgrescluster/datasource_postgres_cluster"
	"github.com/numspot/terraform-provider-numspot/internal/services/postgrescluster/resource_postgres_cluster"
)

const extensionsAttrName = "name"

// extensionsToTfList serializes the read extensions (a *PostgreSQLExtensions named
// slice alias) into the resource model list. The generator's nested resolver
// cannot detect the alias as a slice, so this is wired via a mapping override.
func extensionsToTfList(ctx context.Context, exts *api.PostgreSQLExtensions, diags *diag.Diagnostics) types.List {
	if exts == nil {
		return types.ListNull(resource_postgres_cluster.ExtensionsValue{}.Type(ctx))
	}
	values := make([]resource_postgres_cluster.ExtensionsValue, len(*exts))
	for i, ext := range *exts {
		v, d := resource_postgres_cluster.NewExtensionsValue(resource_postgres_cluster.ExtensionsValue{}.AttributeTypes(ctx), map[string]attr.Value{
			extensionsAttrName: types.StringValue(string(ext.Name)),
		})
		diags.Append(d...)
		values[i] = v
	}
	list, d := types.ListValueFrom(ctx, resource_postgres_cluster.ExtensionsValue{}.Type(ctx), values)
	diags.Append(d...)
	return list
}

// extensionsToDatasourceList is the datasource-model counterpart of
// extensionsToTfList (the datasource ExtensionsValue is a distinct type).
func extensionsToDatasourceList(ctx context.Context, exts *api.PostgreSQLExtensions, diags *diag.Diagnostics) types.List {
	if exts == nil {
		return types.ListNull(datasource_postgres_cluster.ExtensionsValue{}.Type(ctx))
	}
	values := make([]datasource_postgres_cluster.ExtensionsValue, len(*exts))
	for i, ext := range *exts {
		v, d := datasource_postgres_cluster.NewExtensionsValue(
			datasource_postgres_cluster.ExtensionsValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				extensionsAttrName: types.StringValue(string(ext.Name)),
			},
		)
		diags.Append(d...)
		values[i] = v
	}
	list, d := types.ListValueFrom(ctx, datasource_postgres_cluster.ExtensionsValue{}.Type(ctx), values)
	diags.Append(d...)
	return list
}

// extensionsToCreateBody builds the create-body extensions slice from the model
// list. The create body field is a *PostgreSQLExtensions (= *[]PostgreSQLExtension).
func extensionsToCreateBody(ctx context.Context, list types.List, diags *diag.Diagnostics) *api.PostgreSQLExtensions {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var values []resource_postgres_cluster.ExtensionsValue
	diags.Append(list.ElementsAs(ctx, &values, false)...)
	exts := make(api.PostgreSQLExtensions, len(values))
	for i, v := range values {
		exts[i] = api.PostgreSQLExtension{Name: api.PostgreSQLExtensionName(v.Name.ValueString())}
	}
	return &exts
}

// volumeSizeGiB reads the size from whichever union variant the volume carries.
func volumeSizeGiB(v *api.PostgreSQLVolume) int64 {
	if v == nil {
		return 0
	}
	if ev, err := v.AsPostgreSQLEnterpriseVolume(); err == nil {
		return int64(ev.SizeGiB)
	}
	if pv, err := v.AsPostgreSQLPerformanceVolume(); err == nil {
		return int64(pv.SizeGiB)
	}
	if mv, err := v.AsPostgreSQLMagneticVolume(); err == nil {
		return int64(mv.SizeGiB)
	}
	return 0
}

// volumeTypeString reads the storage type from whichever union variant the
// volume carries.
func volumeTypeString(v *api.PostgreSQLVolume) string {
	if v == nil {
		return ""
	}
	if ev, err := v.AsPostgreSQLEnterpriseVolume(); err == nil {
		return string(ev.Type)
	}
	if pv, err := v.AsPostgreSQLPerformanceVolume(); err == nil {
		return string(pv.Type)
	}
	if mv, err := v.AsPostgreSQLMagneticVolume(); err == nil {
		return string(mv.Type)
	}
	return ""
}

// volumeUnionFromTF rebuilds the SDK union from the flat TF size/type pair,
// dispatching on the storage type. Empty/unknown input yields the zero union.
func volumeUnionFromTF(size int64, volumeType string) api.PostgreSQLVolume {
	var out api.PostgreSQLVolume
	switch volumeType {
	case string(api.PERFORMANCE):
		v := api.PostgreSQLPerformanceVolume{SizeGiB: int32(size), Type: api.PERFORMANCE}
		_ = out.FromPostgreSQLPerformanceVolume(v)
	case string(api.MAGNETIC):
		v := api.PostgreSQLMagneticVolume{SizeGiB: int32(size), Type: api.MAGNETIC}
		_ = out.FromPostgreSQLMagneticVolume(v)
	case string(api.ENTERPRISE):
		v := api.PostgreSQLEnterpriseVolume{SizeGiB: int32(size), Type: api.ENTERPRISE}
		_ = out.FromPostgreSQLEnterpriseVolume(v)
	}
	return out
}

// volumeIsEmpty reports whether the union carries no variant (nil JSON body):
// every As* accessor then fails to unmarshal.
func volumeIsEmpty(v *api.PostgreSQLVolume) bool {
	if v == nil {
		return true
	}
	_, errEnt := v.AsPostgreSQLEnterpriseVolume()
	_, errPerf := v.AsPostgreSQLPerformanceVolume()
	_, errMag := v.AsPostgreSQLMagneticVolume()
	return errEnt != nil && errPerf != nil && errMag != nil
}

// serializeVolumeValue converts the read union into the resource-model flat
// nested object.
func serializeVolumeValue(ctx context.Context, v api.PostgreSQLVolume, diags *diag.Diagnostics) resource_postgres_cluster.VolumeValue {
	if volumeIsEmpty(&v) {
		return resource_postgres_cluster.NewVolumeValueNull()
	}
	value, d := resource_postgres_cluster.NewVolumeValue(resource_postgres_cluster.VolumeValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"size_gi_b": types.Int64Value(volumeSizeGiB(&v)),
		"type":      types.StringValue(volumeTypeString(&v)),
	})
	diags.Append(d...)
	return value
}

// volumeToDsObject is the datasource-model counterpart of serializeVolumeValue
// (mappingVolume returns a plain types.Object).
func volumeToDsObject(ctx context.Context, v api.PostgreSQLVolume, diags *diag.Diagnostics) types.Object {
	if volumeIsEmpty(&v) {
		return types.ObjectNull(datasource_postgres_cluster.VolumeValue{}.AttributeTypes(ctx))
	}
	value, d := datasource_postgres_cluster.NewVolumeValue(datasource_postgres_cluster.VolumeValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"size_gi_b": types.Int64Value(volumeSizeGiB(&v)),
		"type":      types.StringValue(volumeTypeString(&v)),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

// deserializeVolumeUnion builds the create-body union from the resource-model
// flat nested object.
func deserializeVolumeUnion(e resource_postgres_cluster.VolumeValue) api.PostgreSQLVolume {
	if e.IsNull() || e.IsUnknown() {
		return api.PostgreSQLVolume{}
	}
	return volumeUnionFromTF(e.SizeGiB.ValueInt64(), e.VolumeType.ValueString())
}
