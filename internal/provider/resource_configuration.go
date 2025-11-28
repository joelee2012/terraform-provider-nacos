package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/joelee2012/nacosctl/pkg/nacos"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ConfigurationResource{}
var _ resource.ResourceWithImportState = &ConfigurationResource{}
var _ resource.ResourceWithIdentity = &ConfigurationResource{}

func NewConfigurationResource() resource.Resource {
	return &ConfigurationResource{}
}

// ConfigurationResource defines the resource implementation.
type ConfigurationResource struct {
	client *nacos.Client
}

// ConfigurationResourceModel describes the resource data model.
type ConfigurationResourceModel struct {
	ID          types.String `tfsdk:"id"`
	DataID      types.String `tfsdk:"data_id"`
	Group       types.String `tfsdk:"group"`
	Content     types.String `tfsdk:"content"`
	NamespaceID types.String `tfsdk:"namespace_id"`
	Type        types.String `tfsdk:"type"`
	Application types.String `tfsdk:"application"`
	Description types.String `tfsdk:"description"`
	Tags        types.Set    `tfsdk:"tags"`
}

func (c *ConfigurationResourceModel) SetFromConfiguration(ctx context.Context, cfg *nacos.Configuration) diag.Diagnostics {
	c.ID = types.StringValue(BuildThreePartID(cfg.GetNamespace(), cfg.GetGroup(), cfg.DataID))
	c.DataID = types.StringValue(cfg.DataID)
	c.Group = types.StringValue(cfg.GetGroup())
	c.NamespaceID = types.StringValue(cfg.GetNamespace())
	c.Application = types.StringValue(cfg.Application)
	c.Content = types.StringValue(cfg.Content)
	c.Description = types.StringValue(cfg.Description)
	c.Type = types.StringValue(cfg.Type)
	var diags diag.Diagnostics
	if cfg.Tags != "" {
		tags, diags := types.SetValueFrom(ctx, types.StringType, strings.Split(cfg.Tags, ","))
		if diags.HasError() {
			return diags
		}
		c.Tags = tags
	}
	return diags
}

func (c *ConfigurationResourceModel) TagsToString(ctx context.Context) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	var tags []string
	elements := make([]types.String, 0, len(c.Tags.Elements()))
	diags.Append(c.Tags.ElementsAs(ctx, &elements, false)...)
	if diags.HasError() {
		return "", diags
	}
	for _, tag := range elements {
		tags = append(tags, tag.ValueString())
	}
	return strings.Join(tags, ","), diags
}

func (r *ConfigurationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configuration"
}

func (r *ConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Nacos configuration resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this Terraform resource. In the format of `<namespace_id>:<group>:<data_id>`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"data_id": schema.StringAttribute{
				MarkdownDescription: "Configuration data id.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "Configuration content.",
				Required:            true,
			},
			"group": schema.StringAttribute{
				MarkdownDescription: "Configuration group, default is `DEFAULT_GROUP`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("DEFAULT_GROUP"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"namespace_id": schema.StringAttribute{
				MarkdownDescription: "Configuration namespace id, default is empty string which means public namespace.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Configuration type, default is `text`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("text"),
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"text", "json", "xml", "yaml", "html", "properties"}...),
				},
			},
			"application": schema.StringAttribute{
				MarkdownDescription: "Configuration application.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Configuration description.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.SetAttribute{
				MarkdownDescription: "Configuration tags.",
				ElementType:         types.StringType,
				Optional:            true,
			},
		},
	}
}

func (r *ConfigurationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigurationResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	getOpts := &nacos.GetCfgOpts{
		DataID:      data.DataID.ValueString(),
		Group:       data.Group.ValueString(),
		NamespaceID: data.NamespaceID.ValueString(),
	}
	tflog.Debug(ctx, "creating configuration", map[string]any{
		"namespace_id": getOpts.NamespaceID,
		"group":        getOpts.Group,
		"data_id":      getOpts.DataID,
	})

	config, err := r.client.GetConfig(getOpts)
	id := BuildThreePartID(getOpts.NamespaceID, getOpts.Group, getOpts.DataID)
	if err == nil && config != nil {
		resp.Diagnostics.AddError(
			"Configuration already exists",
			fmt.Sprintf("A configuration with namespace_id=%s,group=%s,data_id=%s already exists. "+
				"Run `terraform import nacos_configuration.example %s` to manage it.", getOpts.NamespaceID, getOpts.Group, getOpts.DataID, id),
		)
		return
	}
	opts := &nacos.CreateCfgOpts{
		DataID:      data.DataID.ValueString(),
		Group:       data.Group.ValueString(),
		Content:     data.Content.ValueString(),
		NamespaceID: data.NamespaceID.ValueString(),
		Type:        data.Type.ValueString(),
		Application: data.Application.ValueString(),
		Description: data.Description.ValueString(),
	}

	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		tags, diags := data.TagsToString(ctx)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		opts.Tags = tags
	}

	err = r.client.CreateConfig(opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Nacos configuration",
			err.Error(),
		)
		return
	}
	data.ID = types.StringValue(id)

	tflog.Debug(ctx, "created configuration", map[string]any{
		"namespace_id": getOpts.NamespaceID,
		"group":        getOpts.Group,
		"data_id":      getOpts.DataID,
	})
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigurationResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	namespaceId, group, dataId, err := ParseThreePartID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Parse Nacos configuration",
			err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "import configuration", map[string]any{
		"namespace_id": namespaceId,
		"group":        group,
		"data_id":      dataId,
	})
	config, err := r.client.GetConfig(&nacos.GetCfgOpts{
		NamespaceID: namespaceId,
		Group:       group,
		DataID:      dataId,
	})
	if err != nil {
		if IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		}
		resp.Diagnostics.AddError(
			"Unable to Read Nacos configuration",
			err.Error(),
		)
		return
	}

	if config == nil {
		resp.Diagnostics.AddError(
			"No such Nacos configuration",
			data.ID.ValueString(),
		)
		return
	}

	resp.Diagnostics.Append(data.SetFromConfiguration(ctx, config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "found configuration", map[string]any{
		"namespace_id": namespaceId,
		"group":        group,
		"data_id":      dataId,
	})
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigurationResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	opts := &nacos.CreateCfgOpts{
		DataID:      data.DataID.ValueString(),
		Group:       data.Group.ValueString(),
		Content:     data.Content.ValueString(),
		NamespaceID: data.NamespaceID.ValueString(),
		Type:        data.Type.ValueString(),
		Application: data.Application.ValueString(),
		Description: data.Description.ValueString(),
	}

	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		tags, diags := data.TagsToString(ctx)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		opts.Tags = tags
	}
	err := r.client.CreateConfig(opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Nacos namespaces",
			err.Error(),
		)
		return
	}
	config, err := r.client.GetConfig(&nacos.GetCfgOpts{
		DataID:      data.DataID.ValueString(),
		Group:       data.Group.ValueString(),
		NamespaceID: data.NamespaceID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Nacos configuration after updating resource",
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(data.SetFromConfiguration(ctx, config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "updated configuration", map[string]any{
		"namespace_id": opts.NamespaceID,
		"group":        opts.Group,
		"data_id":      opts.DataID,
	})
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConfigurationResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	opts := &nacos.DeleteCfgOpts{
		DataID:      data.DataID.ValueString(),
		Group:       data.Group.ValueString(),
		NamespaceID: data.NamespaceID.ValueString(),
	}
	err := r.client.DeleteConfig(opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Nacos configuration",
			err.Error(),
		)
		return
	}
}

func (r *ConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Struct model for identity data handling.
type ConfigurationResourceIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

func (r *ConfigurationResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true, // must be set during import by the practitioner
			},
		},
	}
}

func BuildThreePartID(namespaceID, group, dataID string) string {
	return fmt.Sprintf("%s:%s:%s", namespaceID, group, dataID)
}

func ParseThreePartID(id string) (namespaceID, group, dataID string, err error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 3 || idParts[1] == "" || idParts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected ID format (%q). expected <namespace_id>:<group>:<data_id>", id)
	}
	return idParts[0], idParts[1], idParts[2], nil
}
