package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/joelee2012/nacosctl/pkg/nacos"
)

func TestAccConfigurationDataSource(t *testing.T) {
	resourceName := "data.nacos_configuration.test"
	dataId := "test-data-id"
	group := "test-group"
	content := `
server:
  url: example.com
  port: 80
`
	namespaceId := ""
	setupTestConfiguration(t, &nacos.CreateCfgOpts{NamespaceID: namespaceId, DataID: dataId, Group: group, Content: content})
	if testClient.APIVersion == "v3" {
		namespaceId = "public"
	}
	config := fmt.Sprintf(`
data "nacos_configuration" "test" {
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
						tfjsonpath.New("namespace_id"),
						knownvalue.StringExact(namespaceId),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("content"),
						knownvalue.StringExact(content),
					),
				},
			},
		},
	})
}
