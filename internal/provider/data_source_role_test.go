package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func testAccRoleDataSourceConfig(name, username string) string {
	return fmt.Sprintf(`
resource "nacos_user" "test" {
	username = "%s"
	password = "123456"
}

resource "nacos_role" "test" {
	username = nacos_user.test.username
	name = "%s"
}

data "nacos_role" "test" {
	username = nacos_role.test.username
	name = nacos_role.test.name
}
`, username, name)
}

func TestAccRoleDataSource(t *testing.T) {
	resourceName := "data.nacos_role.test"
	username := "tf-user-ds"
	name := "tf-role-ds"
	id := fmt.Sprintf("%s:%s", name, username)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccRoleDataSourceConfig(name, username),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("id"),
						knownvalue.StringExact(id),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("username"),
						knownvalue.StringExact(username),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}
