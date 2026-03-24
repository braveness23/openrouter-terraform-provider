# List all available models
data "openrouter_models" "all" {}

output "model_count" {
  value = length(data.openrouter_models.all.models)
}

# List only models that support function calling
data "openrouter_models" "tool_capable" {
  supported_parameters = ["tools"]
}
