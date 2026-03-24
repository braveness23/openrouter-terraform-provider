terraform {
  required_providers {
    openrouter = {
      source  = "braveness23/openrouter"
      version = "~> 0.1"
    }
  }
}

# Configure via environment variables:
#   OPENROUTER_API_KEY            - standard key for model data sources
#   OPENROUTER_MANAGEMENT_API_KEY - management key for resources and credits
provider "openrouter" {
  # api_key            = "sk-or-v1-..."  # or use OPENROUTER_API_KEY env var
  # management_api_key = "sk-or-m-..."  # or use OPENROUTER_MANAGEMENT_API_KEY env var
}
