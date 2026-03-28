# Store a secret as an account-level environment variable
resource "vastai_environment_variable" "hf_token" {
  key   = "HF_TOKEN"
  value = var.huggingface_token
}

variable "huggingface_token" {
  type      = string
  sensitive = true
}
