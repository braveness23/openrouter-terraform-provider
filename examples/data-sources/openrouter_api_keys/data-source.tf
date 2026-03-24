# List all active API keys
data "openrouter_api_keys" "active" {}

# List all keys including disabled ones
data "openrouter_api_keys" "all" {
  include_disabled = true
}

output "active_key_count" {
  value = length(data.openrouter_api_keys.active.keys)
}
