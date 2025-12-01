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
var _ datasource.DataSource = &PermissionDataSource{}

func NewPermissionDataSource() datasource.DataSource {
	return &PermissionDataSource{}
}

// PermissionDataSource defines the data source implementation.
type PermissionDataSource struct {
	client *nacos.Client
}

// PermissionDataSourceModel describes the data source data model.
type PermissionDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	RoleName types.String `tfsdk:"role_name"`
	Resource types.String `tfsdk:"resource"`
	Action   types.String `tfsdk:"action"`
}

func (r *PermissionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission"
}

func (r *PermissionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Nacos permission data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the permission, in the format of `<role_name>:<resource>:<action>`",
				Computed:            true,
			},
			"role_name": schema.StringAttribute{
				MarkdownDescription: "Role name to query",
				Required:            true,
			},
			"resource": schema.StringAttribute{
				MarkdownDescription: "Resource to query, in the format of `<namespace_id>:<group>:<data_id>`",
				Required:            true,
			},
			"action": schema.StringAttribute{
				MarkdownDescription: "Action to query (r, w, rw)",
				Required:            true,
			},
		},
	}
}

func (r *PermissionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*nacos.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *nacos.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *PermissionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PermissionDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	roleName := data.RoleName.ValueString()
	resource := data.Resource.ValueString()
	action := data.Action.ValueString()

	tflog.Debug(ctx, "reading permission", map[string]any{
		"role_name": roleName,
		"resource": resource, 
		"action": action,
	})

	_, err := r.client.GetPermission(roleName, resource, action)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Nacos permission",
			err.Error(),
		)
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s:%s", roleName, resource, action))
	data.RoleName = types.StringValue(roleName)
	data.Resource = types.StringValue(resource)
	data.Action = types.StringValue(action)

	tflog.Debug(ctx, "found permission", map[string]any{"id": data.ID.ValueString()})
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
