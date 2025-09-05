package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/joelee2012/nacosctl/pkg/nacos"
	"github.com/joelee2012/terraform-provider-nacos/internal/provider/testutil"
)

func TestAccConfigurationDataSource(t *testing.T) {
	resourceName := "data.nacos_configuration.test"
	config := `
data "nacos_configuration" "test" {
  data_id = "test-data-id"
  group  = "test-group"
}
`

	testutil.CreateConfiguration(t, &nacos.CreateCfgOpts{NamespaceID: "", DataID: dataId, Group: group, Content: content})

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
						knownvalue.StringExact(""),
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
