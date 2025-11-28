package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/joelee2012/nacosctl/pkg/nacos"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PermissionResource{}
var _ resource.ResourceWithImportState = &PermissionResource{}
var _ resource.ResourceWithIdentity = &PermissionResource{}

func NewPermissionResource() resource.Resource {
	return &PermissionResource{}
}

// PermissionResource defines the resource implementation.
type PermissionResource struct {
	client *nacos.Client
}

// PermissionResourceModel describes the resource data model.
type PermissionResourceModel struct {
	ID       types.String `tfsdk:"id"`
	RoleName types.String `tfsdk:"role_name"`
	Resource types.String `tfsdk:"resource"`
	Action   types.String `tfsdk:"action"`
}

func (r *PermissionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission"
}

func (r *PermissionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Nacos permission resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of this terraform resource, in the format of `<role_name>:<resource>:<action>`",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role_name": schema.StringAttribute{
				MarkdownDescription: "Role name to bind this permission",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource": schema.StringAttribute{
				MarkdownDescription: "Resource to bind this permission, in the format of `<namespace_id>:<group>:<data_id>`",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"action": schema.StringAttribute{
				MarkdownDescription: "Action to bind this permission, choices are `r`, `w`, `rw`",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"r", "w", "rw"}...),
				},
			},
		},
	}
}

func (r *PermissionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func ParesePermissionID(id string) (string, string, string, error) {
	re := regexp.MustCompile(`^([^:]+):([^:]*:[^:]+:[^:]+):(r|w|rw)$`)
	matches := re.FindStringSubmatch(id)
	if matches == nil {
		return "", "", "", fmt.Errorf("unexpected ID format (%q). expected <role_name>:<resource>:<action>", id)
	}
	return matches[1], matches[2], matches[3], nil
}

func (r *PermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PermissionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	rolename := data.RoleName.ValueString()
	resource := data.Resource.ValueString()
	action := data.Action.ValueString()
	id := BuildThreePartID(rolename, resource, action)

	tflog.Debug(ctx, "creating permission", map[string]any{"role_name": rolename, "resource": resource, "action": action})

	perm, err := r.client.GetPermission(rolename, resource, action)
	if err == nil && perm != nil {
		resp.Diagnostics.AddError(
			"Permission already exists",
			fmt.Sprintf("A permission with role_name=%s,resource=%s,action=%s already exists. "+
				"Run `terraform import nacos_permission.example %s` to manage it.", rolename, resource, action, id),
		)
		return
	}

	err = r.client.CreatePermission(rolename, resource, action)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Nacos permission",
			err.Error(),
		)
		return
	}

	data.ID = types.StringValue(id)

	tflog.Debug(ctx, "created role", map[string]any{"role_name": rolename, "resource": resource, "action": action})
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// Set data returned by API in identity
	// identity := PermissionResourceIdentityModel{
	// 	ID: types.StringValue(opts.ID),
	// }
	// resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *PermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PermissionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	id := data.ID.ValueString()
	rolename, resource, action, err := ParesePermissionID(id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Parse Nacos permission",
			err.Error(),
		)
		return
	}
	_, err = r.client.GetPermission(rolename, resource, action)
	if err != nil {
		if IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		}
		resp.Diagnostics.AddError(
			"Unable to Read Nacos permission",
			err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "found permission", map[string]any{"role_name": rolename, "resource": resource, "action": action})

	data.RoleName = types.StringValue(rolename)
	data.Resource = types.StringValue(resource)
	data.Action = types.StringValue(action)
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PermissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PermissionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	rolename := data.RoleName.ValueString()
	resource := data.Resource.ValueString()
	action := data.Action.ValueString()
	err := r.client.DeletePermission(rolename, resource, action)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Nacos permission",
			err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "deleted permission", map[string]any{"role_name": rolename, "resource": resource, "action": action})
}

func (r *PermissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Struct model for identity data handling.
type PermissionResourceIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

func (r *PermissionResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true, // must be set during import by the practitioner
			},
		},
	}
}
