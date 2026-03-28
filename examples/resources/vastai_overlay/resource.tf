# Create an overlay network on top of a cluster
resource "vastai_overlay" "training_net" {
  name       = "training-overlay"
  cluster_id = vastai_cluster.gpu_cluster.id
}
