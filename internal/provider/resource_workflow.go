package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/yu-iskw/terraform-provider-n8n/internal/n8n"
)

var (
	_ resource.Resource                = &workflowResource{}
	_ resource.ResourceWithConfigure   = &workflowResource{}
	_ resource.ResourceWithImportState = &workflowResource{}
)

type workflowResource struct {
	client *n8n.Client
}

func NewWorkflowResource() resource.Resource {
	return &workflowResource{}
}

func (r *workflowResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow"
}

func (r *workflowResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an n8n workflow via the Public API. Provide `nodes`, `connections`, and optional `settings` as JSON strings (typically from `jsonencode(...)` or an editor export).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Workflow ID assigned by n8n.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Workflow name.",
			},
			"nodes": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "JSON array of workflow nodes.",
				Validators:          []validator.String{jsonStringValidator{}},
			},
			"connections": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "JSON object of workflow connections.",
				Validators:          []validator.String{jsonStringValidator{}},
			},
			"settings": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("{}"),
				MarkdownDescription: "JSON object of workflow settings. Defaults to `{}`.",
				Validators:          []validator.String{jsonStringValidator{}},
			},
			"active": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the workflow is active. Create/update writes the document first, then activates or deactivates.",
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

func (r *workflowResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*n8n.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *n8n.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *workflowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workflowModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateWorkflow(ctx, n8n.WorkflowWrite{
		Name:        plan.Name.ValueString(),
		Nodes:       json.RawMessage(plan.Nodes.ValueString()),
		Connections: json.RawMessage(plan.Connections.ValueString()),
		Settings:    rawJSONOrEmptyObject(plan.Settings.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating workflow", err.Error())
		return
	}

	wf, err := r.ensureWorkflowActive(ctx, created.ID, plan.Active.ValueBool(), created.Active)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error activating workflow",
			fmt.Sprintf("Workflow %q was created but active state could not be applied: %v", created.ID, err),
		)
		return
	}
	if wf == nil {
		wf = created
	}

	state := workflowModelFromAPI(wf, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *workflowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workflowModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wf, err := r.client.GetWorkflow(ctx, state.ID.ValueString())
	if err != nil {
		var nf *n8n.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading workflow", err.Error())
		return
	}

	next := workflowModelFromAPI(wf, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &next)...)
}

func (r *workflowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, prior workflowModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var wf *n8n.Workflow
	if workflowDocumentChanged(plan, prior) {
		updated, err := r.client.UpdateWorkflow(ctx, plan.ID.ValueString(), n8n.WorkflowWrite{
			Name:        plan.Name.ValueString(),
			Nodes:       json.RawMessage(plan.Nodes.ValueString()),
			Connections: json.RawMessage(plan.Connections.ValueString()),
			Settings:    rawJSONOrEmptyObject(plan.Settings.ValueString()),
		})
		if err != nil {
			resp.Diagnostics.AddError("Error updating workflow", err.Error())
			return
		}
		wf = updated
		next, err := r.ensureWorkflowActive(ctx, plan.ID.ValueString(), plan.Active.ValueBool(), updated.Active)
		if err != nil {
			resp.Diagnostics.AddError("Error updating workflow active state", err.Error())
			return
		}
		if next != nil {
			wf = next
		}
	} else {
		// Active-only: skip PUT, toggle active, then re-read for complete computed fields.
		if _, err := r.ensureWorkflowActive(ctx, plan.ID.ValueString(), plan.Active.ValueBool(), prior.Active.ValueBool()); err != nil {
			resp.Diagnostics.AddError("Error updating workflow active state", err.Error())
			return
		}
		got, err := r.client.GetWorkflow(ctx, plan.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error reading workflow after update", err.Error())
			return
		}
		wf = got
	}

	state := workflowModelFromAPI(wf, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *workflowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workflowModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteWorkflow(ctx, state.ID.ValueString())
	if err != nil {
		var nf *n8n.NotFoundError
		if errors.As(err, &nf) {
			return
		}
		resp.Diagnostics.AddError("Error deleting workflow", err.Error())
		return
	}
}

func (r *workflowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ensureWorkflowActive activates or deactivates when want differs from current.
// Returns nil, nil when no API call is needed.
func (r *workflowResource) ensureWorkflowActive(ctx context.Context, id string, want, current bool) (*n8n.Workflow, error) {
	if want == current {
		return nil, nil
	}
	if want {
		return r.client.ActivateWorkflow(ctx, id)
	}
	return r.client.DeactivateWorkflow(ctx, id)
}
