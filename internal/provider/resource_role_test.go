package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func testAccRoleSourceConfig(name, username string) string {
	return fmt.Sprintf(`
resource "nacos_user" "test" {
	username = "%s"
	password = "123456"
}
resource "nacos_role" "test" {
  username = nacos_user.test.username
  name = "%s"
}
`, username, name)
}

func TestAccRoleResource(t *testing.T) {
	resourceName := "nacos_role.test"
	username := "tf-user"
	name := "tf-role"
	updatedUsername := "tf-user-updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRoleSourceConfig(name, username),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("id"),
						knownvalue.StringExact(name),
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

			// ImportState testing
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccRoleSourceConfig(name, updatedUsername),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("id"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("username"),
						knownvalue.StringExact(updatedUsername),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
