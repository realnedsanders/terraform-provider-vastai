# Search for network volume offers
data "vastai_network_volume_offers" "nv_offers" {
  limit = 5
}

# Create a network-attached volume for shared storage across instances
resource "vastai_network_volume" "shared_data" {
  offer_id = data.vastai_network_volume_offers.nv_offers.most_affordable.id
  size     = 200
  name     = "shared-model-weights"
}
