package provider_test

import (
	"os"
	"testing"

	"github.com/braveness23/openrouter-terraform-provider/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccProtoV6ProviderFactories is used in every acceptance test.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"openrouter": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// testAccPreCheck validates required environment variables are set before running acceptance tests.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if v := os.Getenv("OPENROUTER_API_KEY"); v == "" {
		t.Fatal("OPENROUTER_API_KEY must be set for acceptance tests")
	}
}

// testAccMgmtPreCheck validates the management key is also set.
func testAccMgmtPreCheck(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)
	if v := os.Getenv("OPENROUTER_MANAGEMENT_API_KEY"); v == "" {
		t.Fatal("OPENROUTER_MANAGEMENT_API_KEY must be set for acceptance tests that manage resources")
	}
}

// providerConfig returns the HCL provider block for tests, reading keys from env vars.
func providerConfig() string {
	return `
provider "openrouter" {}
`
}

// TestAccProvider_basic ensures the provider can be configured and initialized.
func TestAccProvider_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
data "openrouter_models" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.openrouter_models.test", "models.#"),
				),
			},
		},
	})
}
