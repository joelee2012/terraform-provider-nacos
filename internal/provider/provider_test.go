package provider

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/joelee2012/go-nacos"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
// The factory function is called for each Terraform CLI command to create a provider
// server that the CLI can connect to and interact with.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"nacos": providerserver.NewProtocol6WithError(New("test")()),
}

var testClient *nacos.Client

func initTestClient(t *testing.T) {
	if testClient != nil {
		return
	}
	client, err := nacos.NewClient(os.Getenv("NACOS_HOST"), os.Getenv("NACOS_USERNAME"), os.Getenv("NACOS_PASSWORD"))
	if err != nil {
		t.Fatalf("Failed to create Nacos client: %s", err.Error())
	}
	if err := client.Init(context.Background()); err != nil {
		t.Fatalf("Failed to detect Nacos API version: %s", err.Error())
	}
	testClient = client
}

// isV3Server reports whether the test Nacos server exposes the v3 console API
// (Nacos 3.x). It replaces the former testClient.APIVersion field access,
// which became unexported in go-nacos v0.4.0. The test client always
// auto-detects the version, so GetVersion returns the actual server version
// (e.g. "3.1.0").
func isV3Server() bool {
	if testClient == nil {
		return false
	}
	ver, err := testClient.GetVersion(context.Background())
	if err != nil {
		return false
	}
	return strings.HasPrefix(ver, "3")
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("NACOS_HOST"); v == "" {
		t.Fatal("NACOS_HOST must be set for acceptance tests")
	}
	if v := os.Getenv("NACOS_USERNAME"); v == "" {
		t.Fatal("NACOS_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("NACOS_PASSWORD"); v == "" {
		t.Fatal("NACOS_PASSWORD must be set for acceptance tests")
	}
	initTestClient(t)
}

func setupTestConfiguration(t *testing.T, opts *nacos.PublishCfgOpts) {
	if os.Getenv("TF_ACC") == "" {
		return
	}
	initTestClient(t)
	ctx := context.Background()
	if err := testClient.PublishConfig(ctx, opts); err != nil {
		t.Errorf("Error creating %s:%s:%s: %s", opts.NamespaceID, opts.Group, opts.DataID, err.Error())
	}
	t.Cleanup(func() {
		if err := testClient.DeleteConfig(ctx, &nacos.DeleteCfgOpts{NamespaceID: opts.NamespaceID, DataID: opts.DataID, Group: opts.Group}); err != nil {
			t.Errorf("Error deleting %s:%s:%s: %s", opts.NamespaceID, opts.Group, opts.DataID, err.Error())
		}
	})
}
