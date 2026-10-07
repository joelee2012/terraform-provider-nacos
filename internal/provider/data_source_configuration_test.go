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
	setupTestConfiguration(t, &nacos.PublishCfgOpts{NamespaceID: namespaceId, DataID: dataId, Group: group, Content: content})
	// The provider normalizes an empty namespace_id to the server's public
	// namespace id, so the config omits it and the expected state value
	// depends on the server version.
	expectedNamespaceID := ""
	if isV3Server() {
		expectedNamespaceID = "public"
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
						knownvalue.StringExact(expectedNamespaceID),
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
