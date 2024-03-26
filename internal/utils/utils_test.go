package utils

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestFromTfStringToStringPtr(t *testing.T) {
	t.Parallel()
	want := PointerOf("dummy string")
	given := types.StringValue("dummy string")

	got := FromTfStringToStringPtr(given)

	assert.Equal(t, want, got)
}
