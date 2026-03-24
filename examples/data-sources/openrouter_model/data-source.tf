# Look up a specific model by ID
data "openrouter_model" "claude" {
  id = "anthropic/claude-opus-4"
}

output "claude_context_length" {
  value = data.openrouter_model.claude.context_length
}

output "claude_prompt_price" {
  value = data.openrouter_model.claude.prompt_price
}
