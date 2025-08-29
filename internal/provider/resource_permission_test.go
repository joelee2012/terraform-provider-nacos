package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func testAccPermissionSourceConfig(role_name, username, action string) string {
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
  action = "%s"
}
`, username, role_name, action)
}

func TestAccPermissionResource(t *testing.T) {
	resourceName := "nacos_permission.test"
	username := "tf-user"
	role_name := "tf-role"
	action := "r"
	id := fmt.Sprintf("%s:%s:%s", role_name, ":*:*", action)
	updatedAction := "rw"
	idUpdated := fmt.Sprintf("%s:%s:%s", role_name, ":*:*", updatedAction)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPermissionSourceConfig(role_name, username, action),
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
						tfjsonpath.New("action"),
						knownvalue.StringExact(action),
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
				Config: testAccPermissionSourceConfig(role_name, username, updatedAction),
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
						tfjsonpath.New("action"),
						knownvalue.StringExact(updatedAction),
					),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
