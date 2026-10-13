package servercertificate

import (
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services/servercertificate/resource_server_certificate"
)

// deserializeUpdateServerCertificate builds the rename/repath update body. Only
// the fields that actually changed (plan vs state) are set, mirroring the API's
// NewName/NewPath optional semantics. It is referenced from the descriptor's
// core.update.calls args and invoked from the generated Update, which only fires
// when Name or Path differ.
func deserializeUpdateServerCertificate(state, plan *resource_server_certificate.ServerCertificateModel) api.UpdateServerCertificateJSONRequestBody {
	body := api.UpdateServerCertificateJSONRequestBody{}
	if !plan.Name.Equal(state.Name) {
		body.NewName = plan.Name.ValueStringPointer()
	}
	if !plan.Path.Equal(state.Path) {
		body.NewPath = plan.Path.ValueStringPointer()
	}
	return body
}
