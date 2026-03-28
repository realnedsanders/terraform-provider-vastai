# Serverless Endpoint Workflow
#
# This example demonstrates setting up a serverless inference endpoint:
# 1. Create an endpoint with autoscaling configuration
# 2. Add a worker group bound to the endpoint
# 3. Query endpoint status

terraform {
  required_providers {
    vastai = {
      source = "realnedsanders/vastai"
    }
  }
}

provider "vastai" {}

# Step 1: Create a serverless endpoint with autoscaling
resource "vastai_endpoint" "inference" {
  endpoint_name = "my-inference-api"
  min_load      = 0
  target_util   = 0.8
  cold_mult     = 2.0
  cold_workers  = 1
  max_workers   = 5
}

# Step 2: Add a worker group to handle requests
resource "vastai_worker_group" "workers" {
  endpoint_id   = vastai_endpoint.inference.id
  template_hash = "your-template-hash"
  gpu_ram       = 24000
  num_gpus      = 1
}

# Step 3: Query endpoint status
data "vastai_endpoints" "all" {}

output "endpoint_id" {
  value = vastai_endpoint.inference.id
}
