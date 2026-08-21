package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestTypedCredentialMetadataCatalog(t *testing.T) {
	cases := []struct {
		res      resource.Resource
		typeName string
	}{
		{NewCredentialSlackAPIResource(), "n8n_credential_slack_api"},
		{NewCredentialNotionAPIResource(), "n8n_credential_notion_api"},
		{NewCredentialN8nAPIResource(), "n8n_credential_n8n_api"},
		{NewCredentialJiraSoftwareCloudAPIResource(), "n8n_credential_jira_software_cloud_api"},
		{NewCredentialHubspotAppTokenResource(), "n8n_credential_hubspot_app_token"},
		{NewCredentialSerpAPIResource(), "n8n_credential_serp_api"},
		{NewCredentialOpenAIAPIResource(), "n8n_credential_open_ai_api"},
		{NewCredentialJWTAuthResource(), "n8n_credential_jwt_auth"},
		{NewCredentialGmailOAuth2Resource(), "n8n_credential_gmail_oauth2"},
		{NewCredentialGoogleSheetsTriggerOAuth2APIResource(), "n8n_credential_google_sheets_trigger_oauth2_api"},
		{NewCredentialGoogleCalendarOAuth2APIResource(), "n8n_credential_google_calendar_oauth2_api"},
		{NewCredentialGoogleCloudStorageOAuth2APIResource(), "n8n_credential_google_cloud_storage_oauth2_api"},
		{NewCredentialGoogleDriveOAuth2APIResource(), "n8n_credential_google_drive_oauth2_api"},
		{NewCredentialGoogleAPIResource(), "n8n_credential_google_api"},
		{NewCredentialGooglePalmAPIResource(), "n8n_credential_google_palm_api"},
		{NewCredentialMCPOAuth2APIResource(), "n8n_credential_mcp_oauth2_api"},
		{NewCredentialGitHubOAuth2APIResource(), "n8n_credential_github_oauth2_api"},
		{NewCredentialSalesforceOAuth2APIResource(), "n8n_credential_salesforce_oauth2_api"},
		{NewCredentialTwitterOAuth2APIResource(), "n8n_credential_twitter_oauth2_api"},
		{NewCredentialMicrosoftGraphSecurityOAuth2APIResource(), "n8n_credential_microsoft_graph_security_oauth2_api"},
	}
	for _, tc := range cases {
		t.Run(tc.typeName, func(t *testing.T) {
			var resp resource.MetadataResponse
			tc.res.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "n8n"}, &resp)
			if resp.TypeName != tc.typeName {
				t.Fatalf("expected %q, got %q", tc.typeName, resp.TypeName)
			}
			var schemaResp resource.SchemaResponse
			tc.res.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
			if schemaResp.Diagnostics.HasError() {
				t.Fatalf("schema: %v", schemaResp.Diagnostics)
			}
		})
	}
}

func TestTokenBuildData(t *testing.T) {
	slack, err := slackAPISpec().BuildData(typedAttrBag{
		strings: map[string]types.String{
			"access_token":     types.StringValue("xoxb-1"),
			"signature_secret": types.StringValue("sig"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if slack["accessToken"] != "xoxb-1" || slack["signatureSecret"] != "sig" {
		t.Fatalf("slack: %#v", slack)
	}

	notion, err := notionAPISpec().BuildData(typedAttrBag{
		strings: map[string]types.String{"api_key": types.StringValue("ntn")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if notion["apiKey"] != "ntn" {
		t.Fatalf("notion: %#v", notion)
	}

	n8nAPI, err := n8nAPISpec().BuildData(typedAttrBag{
		strings: map[string]types.String{
			"api_key":  types.StringValue("key"),
			"base_url": types.StringValue("https://example.com/api/v1"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if n8nAPI["apiKey"] != "key" || n8nAPI["baseUrl"] != "https://example.com/api/v1" {
		t.Fatalf("n8nApi: %#v", n8nAPI)
	}

	jira, err := jiraSoftwareCloudAPISpec().BuildData(typedAttrBag{
		strings: map[string]types.String{
			"email":     types.StringValue("a@b.com"),
			"api_token": types.StringValue("tok"),
			"domain":    types.StringValue("https://ex.atlassian.net"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if jira["email"] != "a@b.com" || jira["apiToken"] != "tok" || jira["domain"] != "https://ex.atlassian.net" {
		t.Fatalf("jira: %#v", jira)
	}

	hub, err := hubspotAppTokenSpec().BuildData(typedAttrBag{
		strings: map[string]types.String{"app_token": types.StringValue("pat")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if hub["appToken"] != "pat" {
		t.Fatalf("hubspot: %#v", hub)
	}

	serp, err := serpAPISpec().BuildData(typedAttrBag{
		strings: map[string]types.String{"api_key": types.StringValue("serp")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if serp["apiKey"] != "serp" {
		t.Fatalf("serp: %#v", serp)
	}
}

func TestOpenAIAndJWTBuildData(t *testing.T) {
	openAI, err := openAIAPISpec().BuildData(typedAttrBag{
		strings: map[string]types.String{
			"api_key":         types.StringValue("sk"),
			"organization_id": types.StringValue("org"),
			"url":             types.StringValue("https://api.openai.com/v1"),
			"header_name":     types.StringValue("X-Custom"),
			"header_value":    types.StringValue("hv"),
		},
		bools: map[string]types.Bool{"header": types.BoolValue(true)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if openAI["apiKey"] != "sk" || openAI["organizationId"] != "org" || openAI["headerName"] != "X-Custom" || openAI["header"] != true {
		t.Fatalf("openai: %#v", openAI)
	}

	jwt, err := jwtAuthSpec().BuildData(typedAttrBag{
		strings: map[string]types.String{
			"key_type":  types.StringValue("passphrase"),
			"secret":    types.StringValue("s"),
			"algorithm": types.StringValue("HS256"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if jwt["keyType"] != "passphrase" || jwt["secret"] != "s" || jwt["algorithm"] != "HS256" {
		t.Fatalf("jwt: %#v", jwt)
	}
}

func TestGoogleBuildData(t *testing.T) {
	gmail, err := gmailOAuth2Spec().BuildData(typedAttrBag{
		strings: map[string]types.String{
			"client_id":      types.StringValue("cid"),
			"client_secret":  types.StringValue("csec"),
			"enabled_scopes": types.StringValue("scope-a"),
		},
		bools: map[string]types.Bool{"custom_scopes": types.BoolValue(true)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if gmail["clientId"] != "cid" || gmail["clientSecret"] != "csec" || gmail["customScopes"] != true || gmail["enabledScopes"] != "scope-a" {
		t.Fatalf("gmail: %#v", gmail)
	}
	if _, ok := gmail["oauthTokenData"]; ok {
		t.Fatal("oauthTokenData should be omitted when unset")
	}

	sa, err := googleAPISpec().BuildData(typedAttrBag{
		strings: map[string]types.String{
			"email":           types.StringValue("sa@x.iam.gserviceaccount.com"),
			"private_key":     types.StringValue("-----BEGIN PRIVATE KEY-----"),
			"region":          types.StringValue("global"),
			"delegated_email": types.StringValue("user@x.com"),
			"scopes":          types.StringValue("https://www.googleapis.com/auth/drive"),
		},
		bools: map[string]types.Bool{
			"impersonate": types.BoolValue(true),
			"http_node":   types.BoolValue(true),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if sa["inpersonate"] != true {
		t.Fatalf("expected inpersonate key, got %#v", sa)
	}
	if _, ok := sa["impersonate"]; ok {
		t.Fatal("must not send Terraform impersonate key")
	}

	palm, err := googlePalmAPISpec().BuildData(typedAttrBag{
		strings: map[string]types.String{
			"api_key": types.StringValue("gem"),
			"host":    types.StringValue("https://generativelanguage.googleapis.com"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if palm["apiKey"] != "gem" {
		t.Fatalf("palm: %#v", palm)
	}
}

func TestOAuthBuildData(t *testing.T) {
	gh, err := githubOAuth2APISpec().BuildData(typedAttrBag{
		strings: map[string]types.String{
			"client_id":     types.StringValue("cid"),
			"client_secret": types.StringValue("csec"),
			"server":        types.StringValue("https://github.example.com/api/v3"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if gh["server"] != "https://github.example.com/api/v3" || gh["clientId"] != "cid" {
		t.Fatalf("github: %#v", gh)
	}

	sf, err := salesforceOAuth2APISpec().BuildData(typedAttrBag{
		strings: map[string]types.String{
			"client_id":     types.StringValue("cid"),
			"client_secret": types.StringValue("csec"),
			"environment":   types.StringValue("sandbox"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if sf["environment"] != "sandbox" {
		t.Fatalf("salesforce: %#v", sf)
	}

	mcp, err := mcpOAuth2APISpec().BuildData(typedAttrBag{
		bools: map[string]types.Bool{"use_dynamic_client_registration": types.BoolValue(true)},
		strings: map[string]types.String{
			"server_url":   types.StringValue("https://mcp.example"),
			"resource_url": types.StringValue("https://mcp.example/res"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if mcp["useDynamicClientRegistration"] != true || mcp["serverUrl"] != "https://mcp.example" {
		t.Fatalf("mcp: %#v", mcp)
	}

	ms, err := microsoftGraphSecurityOAuth2APISpec().BuildData(typedAttrBag{
		strings: map[string]types.String{
			"client_id":              types.StringValue("cid"),
			"client_credential_type": types.StringValue("certificate"),
			"private_key":            types.StringValue("key"),
			"certificate":            types.StringValue("cert"),
			"graph_api_base_url":     types.StringValue("https://graph.microsoft.com"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if ms["clientCredentialType"] != "certificate" || ms["privateKey"] != "key" {
		t.Fatalf("ms graph: %#v", ms)
	}
}
