# Read the current authenticated user's profile
data "vastai_user" "me" {}

output "username" {
  value = data.vastai_user.me.username
}

output "balance" {
  value = data.vastai_user.me.balance
}
