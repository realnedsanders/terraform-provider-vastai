# Create a serverless inference endpoint with autoscaling
resource "vastai_endpoint" "llm_serving" {
  endpoint_name = "llm-inference-v1"

  # Autoscaling configuration
  min_load     = 0
  target_util  = 0.85
  cold_mult    = 2.0
  cold_workers = 3
  max_workers  = 10
}
