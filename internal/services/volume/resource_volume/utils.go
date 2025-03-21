package resource_volume

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

func ReplaceVolumeSize(ctx context.Context, req planmodifier.Int64Request, resp *int64planmodifier.RequiresReplaceIfFuncResponse) {
	resp.RequiresReplace = false

	var state, plan VolumeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateSize := state.Size.ValueInt64()
	planSize := plan.Size.ValueInt64()
	ReplaceVolumeOnDownsize := plan.ReplaceVolumeOnDownsize.ValueBool()

	if planSize < stateSize && ReplaceVolumeOnDownsize {
		resp.RequiresReplace = true
	}
}
