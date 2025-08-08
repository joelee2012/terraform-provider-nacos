// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccNamespacesDataSource(t *testing.T) {
	resourceName := "data.nacos_namespaces.all"
	namespaceId := ""
	name := "public"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: `data "nacos_namespaces" "all" {}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("namespaces").AtSliceIndex(0).AtMapKey("namespace_id"),
						knownvalue.StringExact(namespaceId),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("namespaces").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}
