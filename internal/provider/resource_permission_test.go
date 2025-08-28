package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func testAccPermissionSourceConfig(role_name, username, permission string) string {
	return fmt.Sprintf(`
resource "nacos_user" "test" {
	username = "%s"
	password = "123456"
}
resource "nacos_role" "test" {
  username = nacos_user.test.username
  name = "%s"
}

resource "nacos_permission" "test" {
  role_name = nacos_role.test.name
  resource = ":*:*"
  permission = "%s"
}
`, username, role_name, permission)
}

func TestAccPermissionResource(t *testing.T) {
	resourceName := "nacos_permission.test"
	username := "tf-user"
	role_name := "tf-role"
	permission := "r"
	id := fmt.Sprintf("%s:%s:%s", role_name, ":*:*", permission)
	updatedPermission := "rw"
	idUpdated := fmt.Sprintf("%s:%s:%s", role_name, ":*:*", updatedPermission)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPermissionSourceConfig(role_name, username, permission),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("id"),
						knownvalue.StringExact(id),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("role_name"),
						knownvalue.StringExact(role_name),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("permission"),
						knownvalue.StringExact(permission),
					),
				},
			},

			// ImportState testing
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccPermissionSourceConfig(role_name, username, updatedPermission),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("id"),
						knownvalue.StringExact(idUpdated),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("role_name"),
						knownvalue.StringExact(role_name),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("permission"),
						knownvalue.StringExact(updatedPermission),
					),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
