package provider

import (
	"context"
	"fmt"
	"os"
	"unicode"

	stringvalidator "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/joelee2012/go-nacos"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &NacosProvider{}

// NacosProvider defines the provider implementation.
type NacosProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// NacosProviderModel describes the provider data model.
type NacosProviderModel struct {
	Host       types.String `tfsdk:"host"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	APIVersion types.String `tfsdk:"api_version"`
}

func (p *NacosProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "nacos"
	resp.Version = p.version
}

func (p *NacosProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Terraform Provider for [Nacos](https://nacos.io/)",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "URL of nacos server, set the value statically in the configuration, or use the `NACOS_HOST` environment variable.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for nacos server, set the value statically in the configuration, or use the `NACOS_USERNAME` environment variable.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for nacos server, set the value statically in the configuration, or use the `NACOS_PASSWORD` environment variable.",
				Optional:            true,
			},
			"api_version": schema.StringAttribute{
				MarkdownDescription: "API version of nacos server (`v1` for Nacos v2.x, `v3` for Nacos v3.x). If not set, the provider will auto-detect the version. Set the value statically in the configuration, or use the `NACOS_API_VERSION` environment variable.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("v1", "v3"),
				},
			},
		},
	}
}

// resolveStringAttr resolves a provider string attribute to its effective
// value: the configured value when set, otherwise the named environment
// variable. When the value is still unknown at plan time it records an
// attribute error and returns "" so the caller can stop after collecting
// errors. The lowercase name (e.g. "host") drives both the error summary
// ("Nacos API Host") and the message body.
func resolveStringAttr(resp *provider.ConfigureResponse, attr path.Path, name, envVar string, v types.String) string {
	if v.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			attr,
			"Unknown Nacos API "+titleCase(name),
			fmt.Sprintf("The provider cannot create the Nacos API client as there is an unknown configuration value for the Nacos API %s. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the %s environment variable.",
				name, envVar),
		)
		return ""
	}
	if !v.IsNull() {
		return v.ValueString()
	}
	return os.Getenv(envVar)
}

// requireStringAttr resolves a required provider string attribute to its
// effective value and reports an error when it cannot be used. An unknown
// value (not yet known at plan time) is reported with guidance to apply its
// source; a resolved empty value is reported as missing. The unknown check
// short-circuits before the empty check, since ValueString of an unknown
// value is "" and would otherwise be misreported as missing. The lowercase
// name (e.g. "host") drives the error summary ("Nacos API Host") and body.
func requireStringAttr(resp *provider.ConfigureResponse, attr path.Path, name, envVar string, v types.String) string {
	if v.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			attr,
			"Unknown Nacos API "+titleCase(name),
			fmt.Sprintf("The provider cannot create the Nacos API client as there is an unknown configuration value for the Nacos API %s. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the %s environment variable.",
				name, envVar),
		)
		return ""
	}
	value := os.Getenv(envVar)
	if !v.IsNull() {
		value = v.ValueString()
	}
	if value == "" {
		resp.Diagnostics.AddAttributeError(
			attr,
			"Missing Nacos API "+titleCase(name),
			fmt.Sprintf("The provider cannot create the Nacos API client as there is a missing or empty value for the Nacos API %s. "+
				"Set the %s value in the configuration or use the %s environment variable. "+
				"If either is already set, ensure the value is not empty.",
				name, name, envVar),
		)
	}
	return value
}

// titleCase capitalizes the first rune of s (used for diagnostic summaries).
func titleCase(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func (p *NacosProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config NacosProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve each string attribute from the configuration, falling back to
	// its environment variable. Required attributes (host/username/password)
	// are validated for unknown/empty in the same call; api_version is
	// optional and auto-detected when absent, so it is resolved without an
	// empty check.
	host := requireStringAttr(resp, path.Root("host"), "host", "NACOS_HOST", config.Host)
	username := requireStringAttr(resp, path.Root("username"), "username", "NACOS_USERNAME", config.Username)
	password := requireStringAttr(resp, path.Root("password"), "password", "NACOS_PASSWORD", config.Password)
	apiVersion := resolveStringAttr(resp, path.Root("api_version"), "version", "NACOS_API_VERSION", config.APIVersion)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new Nacos client using the configuration values. When an API
	// version is configured, pin it via WithAPIVersion to skip the server
	// state probe; otherwise the client auto-detects the version during Init.
	var opts []nacos.Option
	if apiVersion != "" {
		opts = append(opts, nacos.WithAPIVersion(apiVersion))
	}
	client, err := nacos.NewClient(host, username, password, opts...)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Nacos API client",
			err.Error(),
		)
		return
	}
	// Detect (or validate the pinned) API version early so that resource and
	// data-source operations build the correct URL paths from the start.
	if err := client.Init(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Unable to detect Nacos API version",
			err.Error(),
		)
		return
	}
	// Example client configuration for data sources and resources
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *NacosProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewNamespaceResource,
		NewConfigurationResource,
		NewUserResource,
		NewRoleResource,
		NewPermissionResource,
	}
}

func (p *NacosProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewNamespaceDataSource,
		NewNamespacesDataSource,
		NewConfigurationDataSource,
		NewConfigurationsDataSource,
		NewUserDataSource,
		NewRoleDataSource,
		NewPermissionDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &NacosProvider{
			version: version,
		}
	}
}
