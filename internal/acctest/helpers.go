package acctest

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/realnedsanders/terraform-provider-vastai/internal/provider"
)

// ProtoV6ProviderFactories is used by acceptance tests to create provider instances.
// It maps provider names to factory functions that create protocol v6 provider servers.
var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"vastai": providerserver.NewProtocol6WithError(provider.New("test")()),
}
