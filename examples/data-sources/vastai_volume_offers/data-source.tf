# Search for volume storage offers
data "vastai_volume_offers" "affordable" {
  max_storage_cost = 0.10
  limit            = 10
}

output "cheapest_volume_offer" {
  value = data.vastai_volume_offers.affordable.most_affordable.id
}
