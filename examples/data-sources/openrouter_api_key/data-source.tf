# Look up an existing API key by its hash
data "openrouter_api_key" "existing" {
  hash = "abc123def456..."
}

output "key_usage" {
  value = data.openrouter_api_key.existing.usage
}
