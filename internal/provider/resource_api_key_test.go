package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAPIKeyResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccMgmtPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "openrouter_api_key" "test" {
  name = "tf-acc-test-basic"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("openrouter_api_key.test", "id"),
					resource.TestCheckResourceAttrSet("openrouter_api_key.test", "hash"),
					resource.TestCheckResourceAttrSet("openrouter_api_key.test", "key"),
					resource.TestCheckResourceAttr("openrouter_api_key.test", "name", "tf-acc-test-basic"),
					resource.TestCheckResourceAttr("openrouter_api_key.test", "disabled", "false"),
				),
			},
		},
	})
}

func TestAccAPIKeyResource_withLimit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccMgmtPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "openrouter_api_key" "test" {
  name        = "tf-acc-test-limit"
  limit       = 5.00
  limit_reset = "monthly"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("openrouter_api_key.test", "name", "tf-acc-test-limit"),
					resource.TestCheckResourceAttr("openrouter_api_key.test", "limit", "5"),
					resource.TestCheckResourceAttr("openrouter_api_key.test", "limit_reset", "monthly"),
				),
			},
			// Update name — limit should be preserved.
			{
				Config: providerConfig() + `
resource "openrouter_api_key" "test" {
  name        = "tf-acc-test-limit-renamed"
  limit       = 5.00
  limit_reset = "monthly"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("openrouter_api_key.test", "name", "tf-acc-test-limit-renamed"),
					resource.TestCheckResourceAttr("openrouter_api_key.test", "limit", "5"),
				),
			},
		},
	})
}

func TestAccAPIKeyResource_disableUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccMgmtPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "openrouter_api_key" "test" {
  name = "tf-acc-test-disable"
}`,
				Check: resource.TestCheckResourceAttr("openrouter_api_key.test", "disabled", "false"),
			},
			{
				Config: providerConfig() + `
resource "openrouter_api_key" "test" {
  name     = "tf-acc-test-disable"
  disabled = true
}`,
				Check: resource.TestCheckResourceAttr("openrouter_api_key.test", "disabled", "true"),
			},
		},
	})
}

func TestAccAPIKeyResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccMgmtPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "openrouter_api_key" "test" {
  name = "tf-acc-test-import"
}`,
			},
			{
				ResourceName:            "openrouter_api_key.test",
				ImportState:             true,
				ImportStateVerify:       true,
				// key is null after import — it cannot be recovered.
				ImportStateVerifyIgnore: []string{"key"},
			},
		},
	})
}

func TestAccAPIKeyResource_limitClear(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccMgmtPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "openrouter_api_key" "test" {
  name  = "tf-acc-test-limitclear"
  limit = %v
}`, 10.0),
				Check: resource.TestCheckResourceAttr("openrouter_api_key.test", "limit", "10"),
			},
			// Remove the limit (set to null/unset).
			{
				Config: providerConfig() + `
resource "openrouter_api_key" "test" {
  name = "tf-acc-test-limitclear"
}`,
				Check: resource.TestCheckNoResourceAttr("openrouter_api_key.test", "limit"),
			},
		},
	})
}
