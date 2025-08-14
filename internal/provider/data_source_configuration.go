package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/joelee2012/nacosctl/pkg/nacos"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ConfigurationDataSource{}

func NewConfigurationDataSource() datasource.DataSource {
	return &ConfigurationDataSource{}
}

// ConfigurationDataSource defines the data source implementation.
type ConfigurationDataSource struct {
	client *nacos.Client
}

// ConfigurationDataSourceModel describes the data source data model.
type ConfigurationDataSourceModel struct {
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
	Desc             types.String `tfsdk:"description"`
	Tags             types.Set    `tfsdk:"tags"`
}

func (d *ConfigurationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configuration"
}

func (d *ConfigurationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Configuration data source",

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
			"tags": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
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
	}

}

func (d *ConfigurationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ConfigurationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ConfigurationDataSourceModel
	// // Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	config, err := d.client.GetConfig(&nacos.GetCSOpts{DataID: data.DataID.ValueString(), Group: data.Group.ValueString(), NamespaceID: data.NamespaceID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Nacos configuration",
			err.Error(),
		)
		return
	}

	data = ConfigurationDataSourceModel{
		ID:               types.StringValue(BuildThreePartID(config.NamespaceId, config.Group, config.DataID)),
		DataID:           types.StringValue(config.DataID),
		Group:            types.StringValue(config.Group),
		Content:          types.StringValue(config.Content),
		NamespaceID:      types.StringValue(config.NamespaceId),
		Type:             types.StringValue(config.Type),
		Md5:              types.StringValue(config.Md5),
		EncryptedDataKey: types.StringValue(config.EncryptedDataKey),
		Application:      types.StringValue(config.AppName),
		CreateTime:       types.Int64Value(config.CreateTime),
		ModifyTime:       types.Int64Value(config.ModifyTime),
		Desc:             types.StringValue(config.Desc),
	}
	if config.Tags != "" {
		tags, diags := types.SetValueFrom(ctx, types.StringType, strings.Split(config.Tags, ","))
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Tags = tags
	} else {
		data.Tags = types.SetNull(types.StringType)
	}
	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
