package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func testAccConfigurationSourceConfig(namespaceId, group, dataId, content, description string) string {
	return fmt.Sprintf(`
resource "nacos_configuration" "test" {
  namespace_id = "%s"
  group        = "%s"
  data_id      = "%s"
  content      = <<EOT
%s
EOT
  description = "%s"
  tags = ["tag1", "tag2"]
}
`, namespaceId, group, dataId, content, description)
}

func TestAccConfigurationResource(t *testing.T) {
	resourceName := "nacos_configuration.test"
	namespaceId := "test-namespace-id"
	dataId := "test-resource-id"
	group := "test-group"
	description := "Test configuration for acceptance testing"
	updatedDescription := "Test configuration for update testing - updated"
	content := `server:
  url: example.com
  port: 80`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccConfigurationSourceConfig(namespaceId, group, dataId, content, description),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("namespace_id"),
						knownvalue.StringExact(namespaceId),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("data_id"),
						knownvalue.StringExact(dataId),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("group"),
						knownvalue.StringExact(group),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("content"),
						knownvalue.StringFunc(func(v string) error {
							if strings.TrimSpace(v) != strings.TrimSpace(content) {
								return fmt.Errorf("[%s] not equal [%s]", v, content)
							}
							return nil
						}),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("tags"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact("tag1"),
							knownvalue.StringExact("tag2"),
						}),
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
				Config: testAccConfigurationSourceConfig(namespaceId, group, dataId, content, updatedDescription),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("namespace_id"),
						knownvalue.StringExact(namespaceId),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("data_id"),
						knownvalue.StringExact(dataId),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("group"),
						knownvalue.StringExact(group),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("content"),
						knownvalue.StringFunc(func(v string) error {
							if strings.TrimSpace(v) != strings.TrimSpace(content) {
								return fmt.Errorf("[%s] not equal [%s]", v, content)
							}
							return nil
						}),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("description"),
						knownvalue.StringExact(updatedDescription),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("tags"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact("tag1"),
							knownvalue.StringExact("tag2"),
						}),
					),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
