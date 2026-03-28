# Search for network volume offers
data "vastai_network_volume_offers" "available" {
  limit = 10
}

output "cheapest_nv_offer" {
  value = data.vastai_network_volume_offers.available.most_affordable.id
}
