package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccActivityDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccMgmtPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
data "openrouter_activity" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.openrouter_activity.test", "activity.#"),
				),
			},
		},
	})
}

func TestAccActivityDataSource_byDate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccMgmtPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
data "openrouter_activity" "test" {
  date = "2026-03-24"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openrouter_activity.test", "date", "2026-03-24"),
					resource.TestCheckResourceAttrSet("data.openrouter_activity.test", "activity.#"),
				),
			},
		},
	})
}
