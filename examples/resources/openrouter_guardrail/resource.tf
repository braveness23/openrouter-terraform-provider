# A guardrail that restricts usage to specific providers with a monthly budget
resource "openrouter_guardrail" "production" {
  name           = "production-guardrail"
  description    = "Restricts to trusted providers with a monthly budget"
  limit_usd      = 100.00
  reset_interval = "monthly"

  allowed_providers = ["Anthropic", "OpenAI"]
  enforce_zdr       = true
}

# A guardrail that blocks specific models
resource "openrouter_guardrail" "safe_models" {
  name          = "safe-models-only"
  allowed_models = [
    "anthropic/claude-opus-4",
    "openai/gpt-4o",
  ]
}
