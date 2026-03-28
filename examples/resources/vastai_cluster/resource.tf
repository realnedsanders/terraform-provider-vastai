# Create a cluster for private networking between machines
resource "vastai_cluster" "gpu_cluster" {
  subnet     = "10.0.0.0/24"
  manager_id = "12345"
}
