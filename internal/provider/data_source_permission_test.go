package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func testAccPermissionDataSourceConfig(roleName, resourceStr, action string) string {
	return fmt.Sprintf(`
resource "nacos_user" "test" {
	username = "tf-user-perm"
	password = "123456"
}

resource "nacos_role" "test" {
	username = nacos_user.test.username
	name = "%s"
}

resource "nacos_permission" "test" {
	role_name = nacos_role.test.name
	resource = "%s"
	action = "%s"
}

data "nacos_permission" "test" {
	role_name = nacos_permission.test.role_name
	resource = nacos_permission.test.resource
	action = nacos_permission.test.action
}
`, roleName, resourceStr, action)
}

func TestAccPermissionDataSource(t *testing.T) {
	resourceName := "data.nacos_permission.test"
	roleName := "tf-role-perm"
	resourceStr := "test-ns:DEFAULT_GROUP:test-data"
	action := "rw"
	id := fmt.Sprintf("%s:%s:%s", roleName, resourceStr, action)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccPermissionDataSourceConfig(roleName, resourceStr, action),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("id"),
						knownvalue.StringExact(id),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("role_name"),
						knownvalue.StringExact(roleName),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("resource"),
						knownvalue.StringExact(resourceStr),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("action"),
						knownvalue.StringExact(action),
					),
				},
			},
		},
	})
}
