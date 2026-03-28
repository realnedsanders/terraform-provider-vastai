package services

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	// Import all sweeper packages to register their sweepers via init().
	// Packages without sweepers (sshkey, cluster, subaccount) are excluded --
	// see their sweep_test.go files for rationale.
	_ "github.com/realnedsanders/terraform-provider-vastai/internal/services/apikey"
	_ "github.com/realnedsanders/terraform-provider-vastai/internal/services/endpoint"
	_ "github.com/realnedsanders/terraform-provider-vastai/internal/services/envvar"
	_ "github.com/realnedsanders/terraform-provider-vastai/internal/services/instance"
	_ "github.com/realnedsanders/terraform-provider-vastai/internal/services/networkvolume"
	_ "github.com/realnedsanders/terraform-provider-vastai/internal/services/overlay"
	_ "github.com/realnedsanders/terraform-provider-vastai/internal/services/team"
	_ "github.com/realnedsanders/terraform-provider-vastai/internal/services/template"
	_ "github.com/realnedsanders/terraform-provider-vastai/internal/services/volume"
	_ "github.com/realnedsanders/terraform-provider-vastai/internal/services/workergroup"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}
