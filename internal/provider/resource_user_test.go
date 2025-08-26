package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func testAccUserSourceConfig(username, password string) string {
	return fmt.Sprintf(`
resource "nacos_user" "test" {
  username = "%s"
  password = "%s"
}
`, username, password)
}

func TestAccUserResource(t *testing.T) {
	resourceName := "nacos_user.test"
	username := "tf-user"
	password := "123456"
	updatedUsername := "tf-user-updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccUserSourceConfig(username, password),
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
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("password"),
						knownvalue.StringExact(password),
					),
				},
			},

			// ImportState testing
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// ImportStateVerifyIdentifierAttribute: "namespace_id",
				// ImportStateId:                        namespaceId,
			},
			// Update and Read testing
			{
				Config: testAccUserSourceConfig(updatedUsername, password),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("id"),
						knownvalue.StringExact(username),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("name"),
						knownvalue.StringExact(updatedUsername),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("password"),
						knownvalue.StringExact(password),
					),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
