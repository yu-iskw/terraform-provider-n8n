package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// jsonStringValidator validates that a string attribute contains valid JSON.
type jsonStringValidator struct{}

func (v jsonStringValidator) Description(ctx context.Context) string {
	return "value must be valid JSON"
}

func (v jsonStringValidator) MarkdownDescription(ctx context.Context) string {
	return "value must be valid JSON"
}

func (v jsonStringValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	s := req.ConfigValue.ValueString()
	if s == "" {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid JSON", "JSON string cannot be empty")
		return
	}
	if !json.Valid([]byte(s)) {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid JSON", fmt.Sprintf("value is not valid JSON: %q", truncate(s, 80)))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// jsonSemanticEqual reports whether two JSON strings are semantically equal
// (ignoring key order and insignificant whitespace).
func jsonSemanticEqual(a, b string) bool {
	if a == b {
		return true
	}
	var va, vb any
	if err := json.Unmarshal([]byte(a), &va); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &vb); err != nil {
		return false
	}
	ba, err := json.Marshal(va)
	if err != nil {
		return false
	}
	bb, err := json.Marshal(vb)
	if err != nil {
		return false
	}
	return bytes.Equal(ba, bb)
}

// preferConfigJSON keeps the Terraform/config JSON when it is semantically equal
// to the API response, avoiding noisy drift from n8n adding default fields.
func preferConfigJSON(configJSON, apiJSON string) string {
	if configJSON == "" {
		return apiJSON
	}
	if jsonSemanticEqual(configJSON, apiJSON) {
		return configJSON
	}
	return apiJSON
}

func rawJSONOrEmptyObject(s string) json.RawMessage {
	if s == "" {
		return json.RawMessage(`{}`)
	}
	return json.RawMessage(s)
}

func rawJSONString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "{}"
	}
	return string(raw)
}
