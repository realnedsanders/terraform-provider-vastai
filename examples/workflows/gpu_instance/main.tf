# GPU Instance Workflow
#
# This example demonstrates the complete workflow for launching a GPU instance:
# 1. Search for available GPU offers
# 2. Create a template with your Docker image
# 3. Add an SSH key for access
# 4. Launch an instance from the best offer

terraform {
  required_providers {
    vastai = {
      source = "realnedsanders/vastai"
    }
  }
}

provider "vastai" {
  # Set VASTAI_API_KEY environment variable or configure here
  # api_key = "your-api-key"
}

# Step 1: Search for available GPU offers
data "vastai_gpu_offers" "rtx4090" {
  gpu_name       = "RTX_4090"
  num_gpus       = 1
  gpu_ram_gb_min = 24
  max_price      = 0.50
  order_by       = "price"
  limit          = 5
}

# Step 2: Create a template with your Docker image
resource "vastai_template" "pytorch" {
  image       = "pytorch/pytorch:2.1.0-cuda12.1-cudnn8-runtime"
  ssh_direct  = true
  jupyter     = false
  onstart_cmd = "pip install transformers && echo 'Ready!'"

  env = {
    HUGGINGFACE_TOKEN = "your-hf-token"
  }
}

# Step 3: Add an SSH key for secure access
resource "vastai_ssh_key" "mykey" {
  name       = "my-workstation"
  public_key = file("~/.ssh/id_ed25519.pub")
}

# Step 4: Launch an instance from the cheapest matching offer
resource "vastai_instance" "training" {
  offer_id      = data.vastai_gpu_offers.rtx4090.most_affordable.id
  template_hash = vastai_template.pytorch.id
  disk_gb       = 50
  label         = "training-job"
  ssh_key_ids   = [vastai_ssh_key.mykey.id]
}

# Output the instance connection details
output "instance_ip" {
  value = vastai_instance.training.ssh_host
}

output "instance_port" {
  value = vastai_instance.training.ssh_port
}
