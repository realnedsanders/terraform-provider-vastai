# Search for affordable RTX 4090 GPU offers
data "vastai_gpu_offers" "rtx4090" {
  gpu_name           = "RTX 4090"
  num_gpus           = 1
  max_price_per_hour = 0.50
  datacenter_only    = true
  limit              = 10
  order_by           = "dph_total"
}

# Use the most affordable offer to create an instance
output "best_offer_id" {
  value = data.vastai_gpu_offers.rtx4090.most_affordable.id
}

output "best_price" {
  value = data.vastai_gpu_offers.rtx4090.most_affordable.price_per_hour
}
