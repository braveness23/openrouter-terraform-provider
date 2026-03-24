# Create an API key with a monthly spending limit
resource "openrouter_api_key" "app" {
  name        = "my-application"
  limit       = 20.00
  limit_reset = "monthly"
}

# The raw key value is available as a sensitive output immediately after creation.
# It cannot be recovered from OpenRouter after the initial apply.
output "app_api_key" {
  value     = openrouter_api_key.app.key
  sensitive = true
}

# Create a restricted key that expires after 30 days
resource "openrouter_api_key" "temporary" {
  name       = "temporary-access"
  expires_at = "2026-12-31T23:59:59Z"
  limit      = 5.00
}
