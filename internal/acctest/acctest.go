package acctest

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

var TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"numspot": providerserver.NewProtocol6WithError(provider.New("test", true)()),
}
