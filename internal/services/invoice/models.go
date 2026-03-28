package invoice

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// InvoicesDataSourceModel describes the data source data model for vastai_invoices.
type InvoicesDataSourceModel struct {
	StartDate types.String   `tfsdk:"start_date"`
	EndDate   types.String   `tfsdk:"end_date"`
	Limit     types.Int64    `tfsdk:"limit"`
	Type      types.String   `tfsdk:"type"`
	Invoices  []InvoiceModel `tfsdk:"invoices"`
	Total     types.Int64    `tfsdk:"total"`
}

// InvoiceModel describes a single invoice in the data source list (read-only).
type InvoiceModel struct {
	ID          types.String  `tfsdk:"id"`
	Amount      types.Float64 `tfsdk:"amount"`
	Type        types.String  `tfsdk:"type"`
	Description types.String  `tfsdk:"description"`
	Timestamp   types.String  `tfsdk:"timestamp"`
}
