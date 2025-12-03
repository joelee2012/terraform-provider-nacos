package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func testAccUserDataSourceConfig(username string) string {
	return fmt.Sprintf(`
resource "nacos_user" "test" {
	username = "%s"
	password = "123456"
}

data "nacos_user" "test" {
	username = nacos_user.test.username
}
`, username)
}

func TestAccUserDataSource(t *testing.T) {
	resourceName := "data.nacos_user.test"
	username := "tf-user-ds"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccUserDataSourceConfig(username),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("id"),
						knownvalue.StringExact(username),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("username"),
						knownvalue.StringExact(username),
					),
				},
			},
		},
	})
}
