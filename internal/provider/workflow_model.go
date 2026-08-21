package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yu-iskw/terraform-provider-n8n/internal/n8n"
)

// workflowModel is the shared Terraform attribute set for n8n_workflow resource and data source.
type workflowModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Nodes        types.String `tfsdk:"nodes"`
	Connections  types.String `tfsdk:"connections"`
	Settings     types.String `tfsdk:"settings"`
	Active       types.Bool   `tfsdk:"active"`
	VersionID    types.String `tfsdk:"version_id"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
	IsArchived   types.Bool   `tfsdk:"is_archived"`
	TriggerCount types.Int64  `tfsdk:"trigger_count"`
}

// workflowModelFromAPI maps an API workflow into Terraform state.
// When prior is non-nil, JSON string attributes keep config formatting when semantically equal.
func workflowModelFromAPI(wf *n8n.Workflow, prior *workflowModel) workflowModel {
	nodes := rawJSONString(wf.Nodes)
	connections := rawJSONString(wf.Connections)
	settings := rawJSONString(wf.Settings)
	if prior != nil {
		nodes = preferConfigJSON(prior.Nodes.ValueString(), nodes)
		connections = preferConfigJSON(prior.Connections.ValueString(), connections)
		settings = preferConfigJSON(prior.Settings.ValueString(), settings)
	}
	return workflowModel{
		ID:           types.StringValue(wf.ID),
		Name:         types.StringValue(wf.Name),
		Nodes:        types.StringValue(nodes),
		Connections:  types.StringValue(connections),
		Settings:     types.StringValue(settings),
		Active:       types.BoolValue(wf.Active),
		VersionID:    types.StringValue(wf.VersionID),
		CreatedAt:    types.StringValue(wf.CreatedAt),
		UpdatedAt:    types.StringValue(wf.UpdatedAt),
		IsArchived:   types.BoolValue(wf.IsArchived),
		TriggerCount: types.Int64Value(int64(wf.TriggerCount)),
	}
}

func workflowDocumentChanged(plan, state workflowModel) bool {
	return plan.Name.ValueString() != state.Name.ValueString() ||
		!jsonSemanticEqual(plan.Nodes.ValueString(), state.Nodes.ValueString()) ||
		!jsonSemanticEqual(plan.Connections.ValueString(), state.Connections.ValueString()) ||
		!jsonSemanticEqual(plan.Settings.ValueString(), state.Settings.ValueString())
}
