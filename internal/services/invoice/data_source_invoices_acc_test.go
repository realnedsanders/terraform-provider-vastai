package invoice_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/realnedsanders/terraform-provider-vastai/internal/acctest"
)

// TestAccInvoicesDataSource_basic verifies the invoices data source can be read
// without error. The invoice list may be empty for test accounts, which is acceptable.
func TestAccInvoicesDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInvoicesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vastai_invoices.all", "total"),
				),
			},
		},
	})
}

func testAccInvoicesDataSourceConfig() string {
	return `
data "vastai_invoices" "all" {}
`
}
