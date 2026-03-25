terraform {
  required_providers {
    vastai = {
      source = "realnedsanders/vastai"
    }
  }
}

provider "vastai" {
  # api_key can be set via VASTAI_API_KEY environment variable
  # api_key = "your-api-key-here"

  # api_url defaults to https://console.vast.ai
  # api_url = "https://console.vast.ai"
}
