package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/joelee2012/go-nacos"
)

func TestAccConfigurationsDataSource(t *testing.T) {
	resourceName := "data.nacos_configurations.test"
	dataId := "test-data-id"
	group := "test-group"
	content := `
server:
  url: example.com
  port: 80
`
	namespaceId := ""
	setupTestConfiguration(t, &nacos.PublishCfgOpts{NamespaceID: namespaceId, DataID: dataId, Group: group, Content: content})
	// The provider surfaces the public namespace as "" regardless of the
	// server's internal id ("public" on v3), matching the config.
	expectedNamespaceID := ""
	expectedContent := content
	if isV3Server() {
		// v3's list endpoint does not return config content.
		expectedContent = ""
	}

	config := fmt.Sprintf(`
data "nacos_configurations" "test" {
  data_id = "test-data-id"
  group  = "test-group"
  namespace_id = "%s"
}
`, namespaceId)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: config,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("items").AtSliceIndex(0).AtMapKey("namespace_id"),
						knownvalue.StringExact(expectedNamespaceID),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("items").AtSliceIndex(0).AtMapKey("content"),
						knownvalue.StringExact(expectedContent),
					),
				},
			},
		},
	})
}
