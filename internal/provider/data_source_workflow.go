package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/yu-iskw/terraform-provider-n8n/internal/n8n"
)

var (
	_ datasource.DataSource              = &workflowDataSource{}
	_ datasource.DataSourceWithConfigure = &workflowDataSource{}
)

type workflowDataSource struct {
	client *n8n.Client
}

func NewWorkflowDataSource() datasource.DataSource {
	return &workflowDataSource{}
}

func (d *workflowDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow"
}

func (d *workflowDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an n8n workflow by ID from the Public API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Workflow ID.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Workflow name.",
			},
			"nodes": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "JSON array of workflow nodes.",
			},
			"connections": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "JSON object of workflow connections.",
			},
			"settings": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "JSON object of workflow settings.",
			},
			"active": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the workflow is active.",
			},
			"version_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Current workflow version identifier from n8n.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creation timestamp from n8n.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Last update timestamp from n8n.",
			},
			"is_archived": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the workflow is archived.",
			},
			"trigger_count": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of active trigger nodes reported by n8n.",
			},
		},
	}
}

func (d *workflowDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*n8n.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *n8n.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *workflowDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config workflowModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wf, err := d.client.GetWorkflow(ctx, config.ID.ValueString())
	if err != nil {
		var nf *n8n.NotFoundError
		if errors.As(err, &nf) {
			resp.Diagnostics.AddError("Workflow not found", fmt.Sprintf("No workflow found with id %q", config.ID.ValueString()))
			return
		}
		resp.Diagnostics.AddError("Error reading workflow", err.Error())
		return
	}

	state := workflowModelFromAPI(wf, nil)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
