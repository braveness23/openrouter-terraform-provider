# Check current account credit balance
data "openrouter_credits" "account" {}

output "credits_remaining" {
  value = data.openrouter_credits.account.total_credits - data.openrouter_credits.account.total_usage
}
