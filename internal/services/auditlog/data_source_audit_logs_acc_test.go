package auditlog_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/realnedsanders/terraform-provider-vastai/internal/acctest"
)

// TestAccAuditLogsDataSource_basic verifies the audit logs data source can be read
// without error. The audit log list may be empty for test accounts, which is acceptable.
func TestAccAuditLogsDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAuditLogsDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vastai_audit_logs.all", "audit_logs.#"),
				),
			},
		},
	})
}

func testAccAuditLogsDataSourceConfig() string {
	return `
data "vastai_audit_logs" "all" {}
`
}
