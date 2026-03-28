# List billing invoices for the current month
data "vastai_invoices" "recent" {
  start_date = "2026-03-01"
  end_date   = "2026-03-31"
  limit      = 50
}

output "invoice_count" {
  value = length(data.vastai_invoices.recent.invoices)
}
