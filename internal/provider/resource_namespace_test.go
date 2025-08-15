package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func testAccNamespaceSourceConfig(namespaceId, name, description string) string {
	return fmt.Sprintf(`
resource "nacos_namespace" "test" {
  namespace_id = "%s"
  name         = "%s"
  description  = "%s"
}
`, namespaceId, name, description)
}

func TestAccNamespaceResource(t *testing.T) {
	resourceName := "nacos_namespace.test"
	namespaceId := "test-namespace-id"
	name := "test-namespace"
	updatedName := "test-namespace-updated"
	description := "Test namespace for acceptance testing"
	updatedDescription := "Test namespace for update testing - updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccNamespaceSourceConfig(namespaceId, name, description),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("namespace_id"),
						knownvalue.StringExact(namespaceId),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
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
				Config: testAccNamespaceSourceConfig(namespaceId, updatedName, updatedDescription),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("namespace_id"),
						knownvalue.StringExact(namespaceId),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("name"),
						knownvalue.StringExact(updatedName),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("description"),
						knownvalue.StringExact(updatedDescription),
					),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
