# Search for volume storage offers
data "vastai_volume_offers" "storage" {
  limit = 5
}

# Create a persistent local volume from an offer
resource "vastai_volume" "dataset" {
  offer_id = data.vastai_volume_offers.storage.most_affordable.id
  size     = 100
  name     = "training-dataset"
}
