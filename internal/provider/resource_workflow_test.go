package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestWorkflowResourceMetadata(t *testing.T) {
	res := NewWorkflowResource()

	var resp resource.MetadataResponse
	res.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "n8n"}, &resp)

	if resp.TypeName != "n8n_workflow" {
		t.Fatalf("expected n8n_workflow, got %q", resp.TypeName)
	}
}

func TestJSONSemanticEqual(t *testing.T) {
	if !jsonSemanticEqual(`{"a":1,"b":2}`, `{"b":2,"a":1}`) {
		t.Fatal("expected semantically equal objects")
	}
	if jsonSemanticEqual(`{"a":1}`, `{"a":2}`) {
		t.Fatal("expected unequal objects")
	}
	if preferConfigJSON(`{"a":1}`, `{"a":1,"extra":true}`) == `{"a":1}` {
		// not equal — API added field; prefer API
		if preferConfigJSON(`{"a":1}`, `{"a":1}`) != `{"a":1}` {
			t.Fatal("expected prefer config when equal")
		}
	}
	got := preferConfigJSON(`{"b":2,"a":1}`, `{"a":1,"b":2}`)
	if got != `{"b":2,"a":1}` {
		t.Fatalf("expected config JSON preserved, got %q", got)
	}
	got = preferConfigJSON(`{"a":1}`, `{"a":1,"extra":true}`)
	if got != `{"a":1,"extra":true}` {
		t.Fatalf("expected API JSON when unequal, got %q", got)
	}
}
