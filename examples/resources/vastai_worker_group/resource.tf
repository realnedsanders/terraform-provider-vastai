# Create a serverless endpoint (required parent resource)
resource "vastai_endpoint" "llm_serving" {
  endpoint_name = "llm-inference"
}

# Create a worker group bound to the endpoint
resource "vastai_worker_group" "llm_workers" {
  endpoint_id   = vastai_endpoint.llm_serving.id
  template_hash = "abc123def456"

  search_params = "gpu_ram>=24 num_gpus=1 gpu_name=RTX_4090"
  gpu_ram       = 24.0
  test_workers  = 3
}
