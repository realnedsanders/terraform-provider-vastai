# Create a cluster (required parent resource)
resource "vastai_cluster" "gpu_cluster" {
  name = "training-cluster"
}

# Create an overlay network on top of a cluster
resource "vastai_overlay" "training_net" {
  name       = "training-overlay"
  cluster_id = vastai_cluster.gpu_cluster.id
}
