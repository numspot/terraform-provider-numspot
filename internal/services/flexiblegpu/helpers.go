package flexiblegpu

import (
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services/flexiblegpu/resource_flexible_gpu"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

// vmIDFromModel returns the VM id the GPU is attached to, or "" when unset. It is
// threaded into the create/delete/update core calls via descriptor extra_args.
func vmIDFromModel(model *resource_flexible_gpu.FlexibleGpuModel) string {
	if model == nil || model.VmId.IsNull() || model.VmId.IsUnknown() {
		return ""
	}
	return model.VmId.ValueString()
}

// deserializeUpdateFlexibleGpu builds the (attribute-only) update body passed to
// core.UpdateFlexibleGPU from the planned model.
func deserializeUpdateFlexibleGpu(tf *resource_flexible_gpu.FlexibleGpuModel) api.UpdateFlexibleGpuJSONRequestBody {
	return api.UpdateFlexibleGpuJSONRequestBody{
		DeleteOnVmDeletion: utils.FromTfBoolToBoolPtr(tf.DeleteOnVmDeletion),
	}
}
