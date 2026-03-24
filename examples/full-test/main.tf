terraform {
  required_providers {
    openrouter = {
      source  = "registry.terraform.io/braveness23/openrouter"
      version = "0.0.1"
    }
  }
}

provider "openrouter" {
  # Reads OPENROUTER_API_KEY and OPENROUTER_MANAGEMENT_API_KEY from environment
}

# ============================================================================
# Resources
# ============================================================================

resource "openrouter_api_key" "test" {
  name        = "tf-test-key"
  limit       = 1.00
  limit_reset = "monthly"
}

resource "openrouter_guardrail" "test" {
  name           = "tf-test-guardrail"
  description    = "Created by Terraform provider test"
  limit_usd      = 5.00
  reset_interval = "monthly"
  allowed_providers = ["Anthropic", "OpenAI"]
}

# ============================================================================
# Data Sources
# ============================================================================

# Read back the key we just created (by hash)
data "openrouter_api_key" "test" {
  hash       = openrouter_api_key.test.hash
  depends_on = [openrouter_api_key.test]
}

# List all keys
data "openrouter_api_keys" "all" {
  depends_on = [openrouter_api_key.test]
}

# Account credits
data "openrouter_credits" "account" {}

# Activity for a recent completed day
data "openrouter_activity" "recent" {
  date = "2026-03-23"
}

# All available models
data "openrouter_models" "all" {}

# Models filtered to tool-capable only
data "openrouter_models" "tool_capable" {
  supported_parameters = ["tools"]
}

# Single model lookup
data "openrouter_model" "claude" {
  id = "anthropic/claude-opus-4"
}

# ============================================================================
# Outputs
# ============================================================================

output "created_key_hash" {
  value       = openrouter_api_key.test.hash
  description = "Hash of the created API key"
}

output "created_key_value" {
  value       = openrouter_api_key.test.key
  sensitive   = true
  description = "Raw value of the created API key (sensitive)"
}

output "created_key_limit_remaining" {
  value       = openrouter_api_key.test.limit_remaining
  description = "Remaining credit balance on the test key"
}

output "created_guardrail_id" {
  value       = openrouter_guardrail.test.id
  description = "UUID of the created guardrail"
}

output "datasource_key_name" {
  value       = data.openrouter_api_key.test.name
  description = "Name read back via data source (should match resource)"
}

output "total_keys" {
  value       = length(data.openrouter_api_keys.all.keys)
  description = "Total number of API keys in account"
}

output "credits_remaining" {
  value       = data.openrouter_credits.account.total_credits - data.openrouter_credits.account.total_usage
  description = "Remaining account credits in USD"
}

output "yesterday_total_requests" {
  value       = sum([for row in data.openrouter_activity.recent.activity : row.requests])
  description = "Total requests made yesterday"
}

output "yesterday_total_spend" {
  value       = sum([for row in data.openrouter_activity.recent.activity : row.usage])
  description = "Total USD spent yesterday"
}

output "total_models" {
  value       = length(data.openrouter_models.all.models)
  description = "Total number of available models"
}

output "tool_capable_models" {
  value       = length(data.openrouter_models.tool_capable.models)
  description = "Number of models supporting function calling"
}

output "claude_context_length" {
  value       = data.openrouter_model.claude.context_length
  description = "Claude Opus 4 context window size"
}

output "claude_prompt_price" {
  value       = data.openrouter_model.claude.prompt_price
  description = "Claude Opus 4 prompt price per token"
}
