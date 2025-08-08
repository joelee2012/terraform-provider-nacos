// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
var _ datasource.DataSource = &NamespacesDataSource{}

func NewNamespacesDataSource() datasource.DataSource {
	return &NamespacesDataSource{}
}

// NamespacesDataSource defines the data source implementation.
type NamespacesDataSource struct {
	client *nacos.Client
}

// NamespacesDataSourceModel describes the data source data model.
type NamespacesDataSourceModel struct {
	Namespaces []*NamespaceModel `tfsdk:"namespaces"`
}

type NamespaceModel struct {
	NamespaceId types.String `tfsdk:"namespace_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Quota       types.Int64  `tfsdk:"quota"`
	ConfigCount types.Int64  `tfsdk:"config_count"`
	Type        types.Int64  `tfsdk:"type"`
}

func (d *NamespacesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespaces"
}

func (d *NamespacesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Namespaces data source",

		Attributes: map[string]schema.Attribute{
			"namespaces": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"namespace_id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"quota": schema.Int64Attribute{
							Computed: true,
						},
						"type": schema.Int64Attribute{
							Computed: true,
						},
						"config_count": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}

}

func (d *NamespacesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NamespacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NamespacesDataSourceModel
	// // Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	namespaces, err := d.client.ListNamespace()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Nacos namespaces",
			err.Error(),
		)
		return
	}

	for _, ns := range namespaces.Items {
		namespace := NamespaceModel{
			NamespaceId: types.StringValue(ns.ID),
			Name:        types.StringValue(ns.Name),
			Description: types.StringValue(ns.Description),
			Quota:       types.Int64Value(int64(ns.Quota)),
			Type:        types.Int64Value(int64(ns.Type)),
			ConfigCount: types.Int64Value(int64(ns.ConfigCount)),
		}
		data.Namespaces = append(data.Namespaces, &namespace)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
