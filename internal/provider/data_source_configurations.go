package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/joelee2012/go-nacos"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ConfigurationsDataSource{}

func NewConfigurationsDataSource() datasource.DataSource {
	return &ConfigurationsDataSource{}
}

// ConfigurationsDataSource defines the data source implementation.
type ConfigurationsDataSource struct {
	client *nacos.Client
}

// ConfigurationsDataSourceModel describes the data source data model.
type ConfigurationsDataSourceModel struct {
	NamespaceID types.String          `tfsdk:"namespace_id"`
	DataID      types.String          `tfsdk:"data_id"`
	Group       types.String          `tfsdk:"group"`
	Items       []*ConfigurationModel `tfsdk:"items"`
}

type ConfigurationModel struct {
	ID               types.String `tfsdk:"id"`
	DataID           types.String `tfsdk:"data_id"`
	Group            types.String `tfsdk:"group"`
	Content          types.String `tfsdk:"content"`
	NamespaceID      types.String `tfsdk:"namespace_id"`
	Type             types.String `tfsdk:"type"`
	Md5              types.String `tfsdk:"md5"`
	EncryptedDataKey types.String `tfsdk:"encrypt_key"`
	Application      types.String `tfsdk:"application"`
	CreateTime       types.Int64  `tfsdk:"create_time"`
	ModifyTime       types.Int64  `tfsdk:"modify_time"`
	Description      types.String `tfsdk:"description"`
	// Tags             types.Set    `tfsdk:"tags"`
}

func (d *ConfigurationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configurations"
}

func (d *ConfigurationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Configuration data source",

		Attributes: map[string]schema.Attribute{
			"data_id": schema.StringAttribute{
				Optional: true,
			},
			"group": schema.StringAttribute{
				Optional: true,
			},
			"namespace_id": schema.StringAttribute{
				Optional: true,
			},
			"items": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"data_id": schema.StringAttribute{
							MarkdownDescription: "Configuration data id.",
							Required:            true,
						},
						"group": schema.StringAttribute{
							MarkdownDescription: "Configuration group.",
							Required:            true,
						},
						"namespace_id": schema.StringAttribute{
							MarkdownDescription: "Configuration namespace id.",
							Optional:            true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "The ID of this Terraform resource. In the format of `<namespace_id>:<group>:<data_id>`.",
							Computed:            true,
						},
						"content": schema.StringAttribute{
							MarkdownDescription: "Configuration content.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Configuration type.",
							Computed:            true,
						},
						"md5": schema.StringAttribute{
							MarkdownDescription: "Configuration md5.",
							Computed:            true,
						},
						"encrypt_key": schema.StringAttribute{
							MarkdownDescription: "Configuration encrypt key.",
							Computed:            true,
						},
						"application": schema.StringAttribute{
							MarkdownDescription: "Configuration application.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"create_time": schema.Int64Attribute{
							MarkdownDescription: "Configuration created time.",
							Computed:            true,
						},
						"modify_time": schema.Int64Attribute{
							MarkdownDescription: "Configuration modify time.",
							Computed:            true,
						},
					},
				},
			},
		},
	}

}

func (d *ConfigurationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ConfigurationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ConfigurationsDataSourceModel
	// // Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	allCs := new(nacos.ConfigurationList)
	var err error
	if data.DataID.IsNull() && data.Group.IsNull() && data.NamespaceID.IsNull() {
		allCs, err = d.client.ListAllConfig()
	} else if data.DataID.IsNull() {
		allCs, err = d.client.ListConfigInNs(data.NamespaceID.ValueString(), data.Group.ValueString())
	} else {
		allCs, err = d.client.ListConfig(&nacos.ListCfgOpts{DataID: data.DataID.ValueString(), Group: data.Group.ValueString(), NamespaceID: data.NamespaceID.ValueString()})
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Nacos configuration",
			err.Error(),
		)
		return
	}
	for _, cfg := range allCs.Items {
		data.Items = append(data.Items, &ConfigurationModel{
			ID:               types.StringValue(BuildThreePartID(cfg.NamespaceID, cfg.Group, cfg.DataID)),
			DataID:           types.StringValue(cfg.DataID),
			Group:            types.StringValue(cfg.Group),
			Content:          types.StringValue(cfg.Content),
			NamespaceID:      types.StringValue(cfg.NamespaceID),
			Type:             types.StringValue(cfg.Type),
			Md5:              types.StringValue(cfg.Md5),
			EncryptedDataKey: types.StringValue(cfg.EncryptedDataKey),
			Application:      types.StringValue(cfg.Application),
			CreateTime:       types.Int64Value(cfg.CreateTime),
			ModifyTime:       types.Int64Value(cfg.ModifyTime),
			Description:      types.StringValue(cfg.Description),
			// Tags:             types.StringValue(config.Tags),
		})
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
