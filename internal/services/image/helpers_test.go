package image

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services/image/resource_image"
)

// bsuAttrTypes returns the plain object attribute types of the nested bsu
// object (the generated BsuValue is built on the same attribute map).
func bsuAttrTypes(ctx context.Context) map[string]attr.Type {
	return resource_image.BsuValue{}.AttributeTypes(ctx)
}

// blockDeviceMappingsList builds a types.List of BlockDeviceMappingsValue from
// attribute maps, failing the test on construction errors.
func blockDeviceMappingsList(t *testing.T, ctx context.Context, elems ...map[string]attr.Value) types.List {
	t.Helper()

	values := make([]attr.Value, 0, len(elems))
	for _, attrs := range elems {
		values = append(values, resource_image.NewBlockDeviceMappingsValueMust(
			resource_image.BlockDeviceMappingsValue{}.AttributeTypes(ctx),
			attrs,
		))
	}

	list, listDiags := types.ListValueFrom(ctx, resource_image.BlockDeviceMappingsValue{}.Type(ctx), values)
	if listDiags.HasError() {
		t.Fatalf("unable to build block device mappings list: %v", listDiags)
	}
	return list
}

func TestDeserializeBsu(t *testing.T) {
	ctx := context.Background()
	attrTypes := bsuAttrTypes(ctx)

	tests := map[string]struct {
		obj  types.Object
		want *api.BsuToCreate
	}{
		"null object maps to nil": {
			obj:  types.ObjectNull(attrTypes),
			want: nil,
		},
		"unknown object maps to nil": {
			obj:  types.ObjectUnknown(attrTypes),
			want: nil,
		},
		"known attributes map to BsuToCreate": {
			obj: types.ObjectValueMust(attrTypes, map[string]attr.Value{
				"delete_on_vm_deletion": types.BoolValue(true),
				"iops":                  types.Int64Value(1500),
				"snapshot_id":           types.StringValue("snap-0123456789"),
				"volume_size":           types.Int64Value(20),
				"volume_type":           types.StringValue("io1"),
			}),
			want: &api.BsuToCreate{
				DeleteOnVmDeletion: boolPtr(true),
				Iops:               intPtr(1500),
				SnapshotId:         strPtr("snap-0123456789"),
				VolumeSize:         intPtr(20),
				VolumeType:         strPtr("io1"),
			},
		},
		"null and unknown attributes map to nil pointers": {
			obj: types.ObjectValueMust(attrTypes, map[string]attr.Value{
				"delete_on_vm_deletion": types.BoolNull(),
				"iops":                  types.Int64Unknown(),
				"snapshot_id":           types.StringNull(),
				"volume_size":           types.Int64Null(),
				"volume_type":           types.StringUnknown(),
			}),
			want: &api.BsuToCreate{},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := deserializeBsu(tc.obj)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("deserializeBsu() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestDeserializeBlockDeviceMappings(t *testing.T) {
	ctx := context.Background()
	attrTypes := bsuAttrTypes(ctx)

	tests := map[string]struct {
		list    types.List
		want    []api.BlockDeviceMappingImage
		wantNil bool
	}{
		"null list yields an empty slice, not nil": {
			// The create body sends [] when block_device_mappings is unset;
			// lock this behavior in.
			list:    types.ListNull(resource_image.BlockDeviceMappingsValue{}.Type(ctx)),
			want:    []api.BlockDeviceMappingImage{},
			wantNil: false,
		},
		"empty list yields an empty slice": {
			list: blockDeviceMappingsList(t, ctx),
			want: []api.BlockDeviceMappingImage{},
		},
		"known bsu is mapped": {
			list: blockDeviceMappingsList(t, ctx, map[string]attr.Value{
				"bsu": types.ObjectValueMust(attrTypes, map[string]attr.Value{
					"delete_on_vm_deletion": types.BoolValue(true),
					"iops":                  types.Int64Value(1500),
					"snapshot_id":           types.StringValue("snap-0123456789"),
					"volume_size":           types.Int64Value(20),
					"volume_type":           types.StringValue("io1"),
				}),
				"device_name":         types.StringValue("/dev/sda1"),
				"virtual_device_name": types.StringValue("sda1"),
			}),
			want: []api.BlockDeviceMappingImage{
				{
					Bsu: &api.BsuToCreate{
						DeleteOnVmDeletion: boolPtr(true),
						Iops:               intPtr(1500),
						SnapshotId:         strPtr("snap-0123456789"),
						VolumeSize:         intPtr(20),
						VolumeType:         strPtr("io1"),
					},
					DeviceName:        strPtr("/dev/sda1"),
					VirtualDeviceName: strPtr("sda1"),
				},
			},
		},
		"unknown bsu maps to nil Bsu": {
			list: blockDeviceMappingsList(t, ctx, map[string]attr.Value{
				"bsu":                 types.ObjectUnknown(attrTypes),
				"device_name":         types.StringValue("/dev/sdb"),
				"virtual_device_name": types.StringValue("sdb"),
			}),
			want: []api.BlockDeviceMappingImage{
				{
					Bsu:               nil,
					DeviceName:        strPtr("/dev/sdb"),
					VirtualDeviceName: strPtr("sdb"),
				},
			},
		},
		"null bsu maps to nil Bsu": {
			list: blockDeviceMappingsList(t, ctx, map[string]attr.Value{
				"bsu":                 types.ObjectNull(attrTypes),
				"device_name":         types.StringNull(),
				"virtual_device_name": types.StringNull(),
			}),
			want: []api.BlockDeviceMappingImage{
				{
					Bsu:               nil,
					DeviceName:        nil,
					VirtualDeviceName: nil,
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := deserializeBlockDeviceMappings(ctx, tc.list)
			if tc.wantNil {
				if got != nil {
					t.Errorf("deserializeBlockDeviceMappings() = %#v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("deserializeBlockDeviceMappings() = nil, want non-nil %v", tc.want)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("deserializeBlockDeviceMappings() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
func boolPtr(b bool) *bool    { return &b }
