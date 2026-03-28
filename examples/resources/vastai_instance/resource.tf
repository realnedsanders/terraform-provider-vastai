# Search for an affordable RTX 4090 offer
data "vastai_gpu_offers" "rtx4090" {
  gpu_name           = "RTX 4090"
  num_gpus           = 1
  max_price_per_hour = 0.50
  limit              = 5
}

# Create an SSH key for secure access
resource "vastai_ssh_key" "my_key" {
  ssh_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExample user@workstation"
}

# Launch a GPU instance from the best offer
resource "vastai_instance" "training" {
  offer_id = data.vastai_gpu_offers.rtx4090.most_affordable.id
  disk_gb  = 50
  image    = "pytorch/pytorch:2.1.0-cuda12.1-cudnn8-runtime"
  label    = "ml-training-run"

  ssh_key_ids = [vastai_ssh_key.my_key.id]

  env = {
    WANDB_API_KEY = "your-wandb-key"
    PROJECT_NAME  = "my-experiment"
  }

  onstart = "cd /workspace && pip install -r requirements.txt"
}
