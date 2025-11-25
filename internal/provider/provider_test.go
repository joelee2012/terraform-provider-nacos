package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/joelee2012/nacosctl/pkg/nacos"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
// The factory function is called for each Terraform CLI command to create a provider
// server that the CLI can connect to and interact with.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"nacos": providerserver.NewProtocol6WithError(New("test")()),
}

var testClient = nacos.NewClient(os.Getenv("NACOS_HOST"), os.Getenv("NACOS_USERNAME"), os.Getenv("NACOS_PASSWORD"))

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
	if v := os.Getenv("NACOS_HOST"); v == "" {
		t.Fatal("NACOS_HOST must be set for acceptance tests")
	}
	if v := os.Getenv("NACOS_USERNAME"); v == "" {
		t.Fatal("NACOS_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("NACOS_PASSWORD"); v == "" {
		t.Fatal("NACOS_PASSWORD must be set for acceptance tests")
	}

}

func setupTestConfiguration(t *testing.T, opts *nacos.CreateCfgOpts) {
	if err := testClient.CreateConfig(opts); err != nil {
		t.Errorf("Error creating %s:%s:%s: %s", opts.NamespaceID, opts.Group, opts.DataID, err.Error())
	}
	t.Cleanup(func() {
		if err := testClient.DeleteConfig(&nacos.DeleteCfgOpts{NamespaceID: opts.NamespaceID, DataID: opts.DataID, Group: opts.Group}); err != nil {
			t.Errorf("Error deleting %s:%s:%s: %s", opts.NamespaceID, opts.Group, opts.DataID, err.Error())
		}
	})
}
