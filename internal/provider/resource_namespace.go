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
	"github.com/joelee2012/nacosctl/pkg/nacos"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &NamespaceResource{}
var _ resource.ResourceWithImportState = &NamespaceResource{}
var _ resource.ResourceWithIdentity = &NamespaceResource{}

func NewNamespaceResource() resource.Resource {
	return &NamespaceResource{}
}

// NamespaceResource defines the resource implementation.
type NamespaceResource struct {
	client *nacos.Client
}

// NamespaceResourceModel describes the resource data model.
type NamespaceResourceModel struct {
	ID          types.String `tfsdk:"id"`
	NamespaceId types.String `tfsdk:"namespace_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ConfigCount types.Int64  `tfsdk:"config_count"`
	Quota       types.Int64  `tfsdk:"quota"`
	Type        types.Int64  `tfsdk:"type"`
}

func (n *NamespaceResourceModel) SetFromNamespace(ns *nacos.Namespace) {
	n.ID = types.StringValue(ns.ID)
	n.NamespaceId = types.StringValue(ns.ID)
	n.Description = types.StringValue(ns.Description)
	n.ConfigCount = types.Int64Value(int64(ns.ConfigCount))
	n.Name = types.StringValue(ns.Name)
	n.Quota = types.Int64Value(int64(ns.Quota))
	n.Type = types.Int64Value(int64(ns.Type))
}
func (r *NamespaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespace"
}

func (r *NamespaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Nacos namespace resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of namespace and this terraform resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"namespace_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "ID of namespace",
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of namespace.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of namespace.",
				Optional:            true,
			},
			"quota": schema.Int64Attribute{
				MarkdownDescription: "Quota of namespace.",
				Computed:            true,
			},
			"type": schema.Int64Attribute{
				MarkdownDescription: "type of namespace.",
				Computed:            true,
			},
			"config_count": schema.Int64Attribute{
				MarkdownDescription: "Configuration count of namespace.",
				Computed:            true,
			},
		},
	}
}

func (r *NamespaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NamespaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NamespaceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.client.GetNamespace(data.NamespaceId.ValueString())
	if err == nil && config != nil {
		resp.Diagnostics.AddError(
			"Namespace already exists",
			fmt.Sprintf("A namespace with namespace_id=%s already exists. "+
				"Run `terraform import nacos_namespace.example %s` to manage it.", data.NamespaceId.ValueString(), data.NamespaceId.ValueString()),
		)
		return
	}

	opts := &nacos.CreateNSOpts{
		ID:          data.NamespaceId.ValueString(),
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
	}

	err = r.client.CreateNamespace(opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Nacos namespaces",
			err.Error(),
		)
		return
	}

	ns, err := r.client.GetNamespace(opts.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Nacos namespace after creating resource",
			err.Error(),
		)
		return
	}

	data.SetFromNamespace(ns)
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// Set data returned by API in identity
	// identity := NamespaceResourceIdentityModel{
	// 	ID: types.StringValue(opts.ID),
	// }
	// resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func IsNotFoundError(err error) bool {
	return strings.HasPrefix(err.Error(), "404 Not Found")
}

func (r *NamespaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NamespaceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ns, err := r.client.GetNamespace(data.ID.ValueString())
	if err != nil {
		if IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		}
		resp.Diagnostics.AddError(
			"Unable to Read Nacos namespaces",
			err.Error(),
		)
		return
	}

	data.SetFromNamespace(ns)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NamespaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NamespaceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	opts := &nacos.CreateNSOpts{
		ID:          data.NamespaceId.ValueString(),
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
	}
	err := r.client.UpdateNamespace(opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Nacos namespaces",
			err.Error(),
		)
		return
	}

	ns, err := r.client.GetNamespace(opts.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Nacos namespace after updating resource",
			err.Error(),
		)
		return
	}

	data.SetFromNamespace(ns)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NamespaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NamespaceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNamespace(data.NamespaceId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Nacos namespaces",
			err.Error(),
		)
		return
	}
}

func (r *NamespaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Struct model for identity data handling.
type NamespaceResourceIdentityModel struct {
	ID types.String `tfsdk:"namespace_id"`
}

func (r *NamespaceResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"namespace_id": identityschema.StringAttribute{
				RequiredForImport: true, // must be set during import by the practitioner
			},
		},
	}
}
