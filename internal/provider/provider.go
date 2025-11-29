package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/joelee2012/nacosctl/pkg/nacos"
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
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
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
		},
	}
}

func (p *NacosProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config NacosProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}
	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Nacos API Host",
			"The provider cannot create the Nacos API client as there is an unknown configuration value for the Nacos API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NACOS_HOST environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Nacos API Username",
			"The provider cannot create the Nacos API client as there is an unknown configuration value for the Nacos API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NACOS_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Nacos API Password",
			"The provider cannot create the Nacos API client as there is an unknown configuration value for the Nacos API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NACOS_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("NACOS_HOST")
	username := os.Getenv("NACOS_USERNAME")
	password := os.Getenv("NACOS_PASSWORD")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Nacos API Host",
			"The provider cannot create the Nacos API client as there is a missing or empty value for the Nacos API host. "+
				"Set the host value in the configuration or use the NACOS_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing Nacos API Username",
			"The provider cannot create the Nacos API client as there is a missing or empty value for the Nacos API username. "+
				"Set the username value in the configuration or use the NACOS_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing Nacos API Password",
			"The provider cannot create the Nacos API client as there is a missing or empty value for the Nacos API password. "+
				"Set the password value in the configuration or use the NACOS_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new HashiCups client using the configuration values
	client := nacos.NewClient(host, username, password)
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
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &NacosProvider{
			version: version,
		}
	}
}
