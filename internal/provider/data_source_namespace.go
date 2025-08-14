package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/joelee2012/nacosctl/pkg/nacos"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &NamespaceDataSource{}

func NewNamespaceDataSource() datasource.DataSource {
	return &NamespaceDataSource{}
}

// NamespaceDataSource defines the data source implementation.
type NamespaceDataSource struct {
	client *nacos.Client
}

type NamespaceDataSourceModel struct {
	NamespaceId types.String `tfsdk:"namespace_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Quota       types.Int64  `tfsdk:"quota"`
	ConfigCount types.Int64  `tfsdk:"config_count"`
	Type        types.Int64  `tfsdk:"type"`
}

func (d *NamespaceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespace"
}

func (d *NamespaceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Namespace data source",

		Attributes: map[string]schema.Attribute{
			"namespace_id": schema.StringAttribute{
				MarkdownDescription: "ID of namespace and this terraform resource.",
				Required: true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of namespace.",
				Computed: true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of namespace.",
				Computed: true,
			},
			"quota": schema.Int64Attribute{
				MarkdownDescription: "Quota of namespace.",
				Computed: true,
			},
			"type": schema.Int64Attribute{
				MarkdownDescription: "type of namespace.",
				Computed: true,
			},
			"config_count": schema.Int64Attribute{
				MarkdownDescription: "Configuration count of namespace.",
				Computed: true,
			},
		},
	}

}

func (d *NamespaceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*nacos.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *NamespaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NamespaceDataSourceModel
	// // Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	ns, err := d.client.GetNamespace(data.NamespaceId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Nacos namespaces",
			err.Error(),
		)
		return
	}

	data = NamespaceDataSourceModel{
		NamespaceId: types.StringValue(ns.ID),
		Name:        types.StringValue(ns.Name),
		Description: types.StringValue(ns.Description),
		Quota:       types.Int64Value(int64(ns.Quota)),
		Type:        types.Int64Value(int64(ns.Type)),
		ConfigCount: types.Int64Value(int64(ns.ConfigCount)),
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
