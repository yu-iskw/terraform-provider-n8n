package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestWorkflowDataSourceMetadata(t *testing.T) {
	ds := NewWorkflowDataSource()

	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "n8n"}, &resp)

	if resp.TypeName != "n8n_workflow" {
		t.Fatalf("expected n8n_workflow, got %q", resp.TypeName)
	}
}
