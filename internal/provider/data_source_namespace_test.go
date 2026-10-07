package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccNamespaceDataSource(t *testing.T) {
	resourceName := "data.nacos_namespace.public"
	name := "public"
	// The provider normalizes an empty namespace_id to the server's public
	// namespace id ("" on v1/v2, "public" on v3), so the config always
	// omits it and the expected state value depends on the server version.
	expectedNamespaceID := ""
	if isV3Server() {
		expectedNamespaceID = "public"
	}
	config := fmt.Sprintf(`
data "nacos_namespace" "public" {
	namespace_id = "%s"
}`, "")

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
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}
