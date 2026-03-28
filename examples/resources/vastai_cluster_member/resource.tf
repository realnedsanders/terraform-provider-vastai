# Create a cluster (required parent resource)
resource "vastai_cluster" "gpu_cluster" {
  name = "training-cluster"
}

# Add a machine to an existing cluster
resource "vastai_cluster_member" "worker_node" {
  cluster_id = vastai_cluster.gpu_cluster.id
  machine_id = "67890"
}
