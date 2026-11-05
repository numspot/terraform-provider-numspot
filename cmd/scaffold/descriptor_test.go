package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDescriptorValidate covers the core.update validation gaps: dead
// when_changed/unless_changed overlaps inside a single call and orphan
// read_if_unchanged settings that would generate a serialize of a nil
// response.
func TestDescriptorValidate(t *testing.T) {
	t.Parallel()

	// baseDescriptor returns a descriptor that passes validateRequiredFields,
	// validateScope and validateResourceOrDatasource.
	baseDescriptor := func() *Descriptor {
		return &Descriptor{
			Name:        "vm",
			GoName:      "VM",
			Singular:    "Vm",
			Plural:      "VMs",
			Scope:       "space",
			HasResource: true,
			HasUpdate:   true,
		}
	}

	tests := []struct {
		name       string
		descriptor *Descriptor
		wantErr    error
	}{
		{
			name:       "minimal valid descriptor",
			descriptor: baseDescriptor(),
		},
		{
			name: "update calls with sibling unless_changed are valid",
			descriptor: func() *Descriptor {
				d := baseDescriptor()
				d.HasTags = true
				d.Core.Update.Calls = []CoreUpdateCall{
					{
						FunctionName:  "UpdateVMSecurityGroups",
						WhenChanged:   []string{"SecurityGroupIds"},
						UnlessChanged: []string{"Type", "UserData"},
					},
					{
						FunctionName: "UpdateVMAttributes",
						WhenChanged:  []string{"Type", "UserData"},
					},
				}
				return d
			}(),
		},
		{
			name: "field in both when_changed and unless_changed of the same call is rejected",
			descriptor: func() *Descriptor {
				d := baseDescriptor()
				d.Core.Update.Calls = []CoreUpdateCall{
					{
						FunctionName:  "UpdateVMAttributes",
						WhenChanged:   []string{"Type"},
						UnlessChanged: []string{"Type"},
					},
				}
				return d
			}(),
			wantErr: errDescriptorUpdateCallFieldOverlap,
		},
		{
			name: "unless_changed field missing from sibling when_changed is rejected",
			descriptor: func() *Descriptor {
				d := baseDescriptor()
				d.Core.Update.Calls = []CoreUpdateCall{
					{
						FunctionName:  "UpdateVMSecurityGroups",
						WhenChanged:   []string{"SecurityGroupIds"},
						UnlessChanged: []string{"Type"},
					},
				}
				return d
			}(),
			wantErr: errDescriptorUpdateCallUnlessChanged,
		},
		{
			name: "read_if_unchanged without calls or has_tags is rejected",
			descriptor: func() *Descriptor {
				d := baseDescriptor()
				d.HasUpdate = false
				d.Core.Update.ReadIfUnchanged = true
				return d
			}(),
			wantErr: errDescriptorReadIfUnchangedWithoutUpdate,
		},
		{
			name: "read_if_unchanged with has_tags only is valid",
			descriptor: func() *Descriptor {
				d := baseDescriptor()
				d.HasUpdate = false
				d.HasTags = true
				d.Core.Update.ReadIfUnchanged = true
				return d
			}(),
		},
		{
			name: "read_if_unchanged with calls only is valid",
			descriptor: func() *Descriptor {
				d := baseDescriptor()
				d.Core.Update.ReadIfUnchanged = true
				d.Core.Update.Calls = []CoreUpdateCall{
					{
						FunctionName: "UpdateVPCAttributes",
						WhenChanged:  []string{"Name"},
					},
				}
				return d
			}(),
		},
		{
			name: "read_if_unchanged with calls and has_tags is valid",
			descriptor: func() *Descriptor {
				d := baseDescriptor()
				d.HasTags = true
				d.Core.Update.ReadIfUnchanged = true
				d.Core.Update.Calls = []CoreUpdateCall{
					{
						FunctionName: "UpdateVMAttributes",
						WhenChanged:  []string{"Type"},
					},
				}
				return d
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.descriptor.Validate()

			if tt.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestDescriptorValidateUnknownSafeOverlapErrorMessage(t *testing.T) {
	t.Parallel()

	d := &Descriptor{
		Name:        "vm",
		GoName:      "VM",
		Singular:    "Vm",
		Plural:      "VMs",
		Scope:       "space",
		HasResource: true,
		HasUpdate:   true,
		Core: CoreConfig{
			Update: CoreUpdateConfig{
				Calls: []CoreUpdateCall{
					{
						FunctionName:  "UpdateVMAttributes",
						WhenChanged:   []string{"Type", "UserData"},
						UnlessChanged: []string{"UserData"},
					},
				},
			},
		},
	}

	err := d.Validate()

	assert.ErrorIs(t, err, errDescriptorUpdateCallFieldOverlap)
	assert.ErrorContains(t, err, "UserData")
}
