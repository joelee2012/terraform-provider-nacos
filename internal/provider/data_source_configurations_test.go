package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/joelee2012/nacosctl/pkg/nacos"
)

func TestAccConfigurationsDataSource(t *testing.T) {
	resourceName := "data.nacos_configurations.test"
	config := `
data "nacos_configurations" "test" {
  data_id = "test-data-id"
  group  = "test-group"
}
`

	expect_id := ""
	expect_content := content
	setupTestConfiguration(t, &nacos.CreateCfgOpts{NamespaceID: "", DataID: dataId, Group: group, Content: content})
	if testClient.APIVersion == "v3" {
		expect_id = "public"
		expect_content = ""
	}
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
						knownvalue.StringExact(expect_id),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("items").AtSliceIndex(0).AtMapKey("content"),
						knownvalue.StringExact(expect_content),
					),
				},
			},
		},
	})
}
