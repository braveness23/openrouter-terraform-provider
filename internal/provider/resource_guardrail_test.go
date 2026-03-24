package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGuardrailResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccMgmtPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "openrouter_guardrail" "test" {
  name = "tf-acc-test-guardrail"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("openrouter_guardrail.test", "id"),
					resource.TestCheckResourceAttr("openrouter_guardrail.test", "name", "tf-acc-test-guardrail"),
					resource.TestCheckResourceAttrSet("openrouter_guardrail.test", "created_at"),
				),
			},
		},
	})
}

func TestAccGuardrailResource_withProviders(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccMgmtPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "openrouter_guardrail" "test" {
  name              = "tf-acc-test-guardrail-providers"
  allowed_providers = ["Anthropic", "OpenAI"]
  enforce_zdr       = true
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("openrouter_guardrail.test", "allowed_providers.#", "2"),
					resource.TestCheckResourceAttr("openrouter_guardrail.test", "enforce_zdr", "true"),
				),
			},
		},
	})
}

func TestAccGuardrailResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccMgmtPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "openrouter_guardrail" "test" {
  name = "tf-acc-test-guardrail-update"
}`,
				Check: resource.TestCheckResourceAttr("openrouter_guardrail.test", "name", "tf-acc-test-guardrail-update"),
			},
			{
				Config: providerConfig() + `
resource "openrouter_guardrail" "test" {
  name        = "tf-acc-test-guardrail-update"
  description = "updated description"
  limit_usd   = 25.0
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("openrouter_guardrail.test", "description", "updated description"),
					resource.TestCheckResourceAttr("openrouter_guardrail.test", "limit_usd", "25"),
				),
			},
		},
	})
}

func TestAccGuardrailResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccMgmtPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "openrouter_guardrail" "test" {
  name = "tf-acc-test-guardrail-import"
}`,
			},
			{
				ResourceName:      "openrouter_guardrail.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
