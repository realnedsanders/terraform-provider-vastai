# Create a cluster (required for overlay)
resource "vastai_cluster" "gpu_cluster" {
  name = "training-cluster"
}

# Create an overlay network
resource "vastai_overlay" "training_net" {
  name       = "training-overlay"
  cluster_id = vastai_cluster.gpu_cluster.id
}

# Create an instance to join to the overlay
data "vastai_gpu_offers" "cheapest" {
  num_gpus           = 1
  max_price_per_hour = 0.50
  limit              = 1
}

resource "vastai_instance" "training" {
  offer_id = data.vastai_gpu_offers.cheapest.offers[0].id
  image    = "pytorch/pytorch:latest"
  disk_gb  = 20
}

# Join an instance to an overlay network
# NOTE: Destroying this resource removes it from Terraform state only;
# individual instance removal from overlays is not supported by the API.
resource "vastai_overlay_member" "training_instance" {
  overlay_name = vastai_overlay.training_net.name
  instance_id  = vastai_instance.training.id
}
