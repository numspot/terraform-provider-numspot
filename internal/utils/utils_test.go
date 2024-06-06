package utils

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

func TestFromTfStringToStringPtr(t *testing.T) {
	t.Parallel()
	want := PointerOf("dummy string")
	given := types.StringValue("dummy string")

	got := FromTfStringToStringPtr(given)

	assert.Equal(t, want, got)
}

type StringObjValuable struct {
	value string
}

func (s StringObjValuable) Type(ctx context.Context) attr.Type {
	return nil
}

func (s StringObjValuable) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	// TODO implement me
	// return nil, nil
	panic("implement me")
}

func (s StringObjValuable) Equal(value attr.Value) bool {
	return s.String() == value.String()
}

func (s StringObjValuable) IsNull() bool {
	return false
}

func (s StringObjValuable) IsUnknown() bool {
	return false
}

func (s StringObjValuable) String() string {
	return s.value
}

func (s StringObjValuable) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	panic("implement me")
}

var _ basetypes.ObjectValuable = &StringObjValuable{}

func TestDiff_Create(t *testing.T) {
	t.Parallel()

	givenCurrent := []basetypes.ObjectValuable{}
	givenDesired := []basetypes.ObjectValuable{StringObjValuable{value: "toto"}}

	wantCreate := givenDesired
	wantDelete := givenCurrent

	toCreate, toDelete := Diff(givenCurrent, givenDesired)

	assert.Equal(t, wantCreate, toCreate)
	assert.Equal(t, wantDelete, toDelete)
}

func TestDiff_Update(t *testing.T) {
	t.Parallel()

	givenCurrent := []basetypes.ObjectValuable{StringObjValuable{value: "tata"}}
	givenDesired := []basetypes.ObjectValuable{StringObjValuable{value: "toto"}}

	wantCreate := givenDesired
	wantDelete := givenCurrent

	toCreate, toDelete := Diff(givenCurrent, givenDesired)

	assert.Equal(t, wantCreate, toCreate)
	assert.Equal(t, wantDelete, toDelete)
}

func TestDiff_Update2(t *testing.T) {
	t.Parallel()

	givenCurrent := []basetypes.ObjectValuable{StringObjValuable{value: "tata"}}
	givenDesired := []basetypes.ObjectValuable{StringObjValuable{value: "tata"}, StringObjValuable{value: "toto"}}

	wantCreate := []basetypes.ObjectValuable{StringObjValuable{value: "toto"}}
	wantDelete := []basetypes.ObjectValuable{}

	toCreate, toDelete := Diff(givenCurrent, givenDesired)

	assert.Equal(t, wantCreate, toCreate)
	assert.Equal(t, wantDelete, toDelete)
}

func TestDiff_Delete(t *testing.T) {
	t.Parallel()

	givenCurrent := []basetypes.ObjectValuable{StringObjValuable{value: "tata"}}
	givenDesired := []basetypes.ObjectValuable{}

	wantCreate := []basetypes.ObjectValuable{}
	wantDelete := []basetypes.ObjectValuable{StringObjValuable{value: "tata"}}

	toCreate, toDelete := Diff(givenCurrent, givenDesired)

	assert.Equal(t, wantCreate, toCreate)
	assert.Equal(t, wantDelete, toDelete)
}
