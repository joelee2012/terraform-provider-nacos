package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/joelee2012/nacosctl/pkg/nacos"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &RoleResource{}
var _ resource.ResourceWithImportState = &RoleResource{}
var _ resource.ResourceWithIdentity = &RoleResource{}

func NewRoleResource() resource.Resource {
	return &RoleResource{}
}

// RoleResource defines the resource implementation.
type RoleResource struct {
	client *nacos.Client
}

// RoleResourceModel describes the resource data model.
type RoleResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Username types.String `tfsdk:"username"`
}

func (r *RoleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *RoleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Nacos role resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID this terraform resource, In the format of `<name>:<username>`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				MarkdownDescription: "name of user to bind this role",
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "name of role.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *RoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*nacos.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func BuildRoleID(name, username string) string {
	return fmt.Sprintf("%s:%s", name, username)
}

func ParseRoleID(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected ID format (%q). expected <role_name>:<username>", id)
	}
	return parts[0], parts[1], nil
}
func (r *RoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RoleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	name := data.Name.ValueString()
	username := data.Username.ValueString()

	tflog.Debug(ctx, "creating role", map[string]any{"name": name, "username": username})

	role, err := r.client.GetRole(name, username)
	id := BuildRoleID(name, username)
	if err == nil && role != nil {
		resp.Diagnostics.AddError(
			"Role already exists",
			fmt.Sprintf("A role with name=%s,username=%s already exists. "+
				"Run `terraform import nacos_role.example %s` to manage it.", name, username, id),
		)
		return
	}

	err = r.client.CreateRole(name, username)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Nacos role",
			err.Error(),
		)
		return
	}

	data.ID = types.StringValue(id)

	tflog.Debug(ctx, "created role", map[string]any{"name": name, "username": username})
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// Set data returned by API in identity
	// identity := RoleResourceIdentityModel{
	// 	ID: types.StringValue(opts.ID),
	// }
	// resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *RoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RoleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	id := data.ID.ValueString()
	name, username, err := ParseRoleID(id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Parse Nacos role",
			err.Error(),
		)
		return
	}
	role, err := r.client.GetRole(name, username)
	if err != nil {
		if IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		}
		resp.Diagnostics.AddError(
			"Unable to Read Nacos role",
			err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "found role", map[string]any{"name": name, "username": username})
	data.Name = types.StringValue(role.Name)
	data.Username = types.StringValue(role.Username)
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RoleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	username := data.Username.ValueString()
	tflog.Debug(ctx, "deleting role", map[string]any{"name": name, "username": username})
	err := r.client.DeleteRole(name, username)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Nacos role",
			err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "deleted role", map[string]any{"name": name, "username": username})
}

func (r *RoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Struct model for identity data handling.
type RoleResourceIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

func (r *RoleResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true, // must be set during import by the practitioner
			},
		},
	}
}
