package testutil

import (
	"os"
	"testing"

	"github.com/joelee2012/go-nacos"
)

var client = nacos.NewClient(os.Getenv("NACOS_HOST"), os.Getenv("NACOS_USERNAME"), os.Getenv("NACOS_PASSWORD"))

func CreateConfiguration(t *testing.T, opts *nacos.CreateCfgOpts) {
	if err := client.CreateConfig(opts); err != nil {
		t.Errorf("Error creating %s:%s:%s: %s", opts.NamespaceID, opts.Group, opts.DataID, err.Error())
	}
	t.Cleanup(func() {
		if err := client.DeleteConfig(&nacos.DeleteCfgOpts{NamespaceID: opts.NamespaceID, DataID: opts.DataID, Group: opts.Group}); err != nil {
			t.Errorf("Error deleting %s:%s:%s: %s", opts.NamespaceID, opts.Group, opts.DataID, err.Error())
		}
	})
}
