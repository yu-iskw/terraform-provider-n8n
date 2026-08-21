package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccN8nWorkflow_basic(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test unless TF_ACC=1")
	}

	name := fmt.Sprintf("tf-acc-workflow-%d", os.Getpid())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig(name, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_workflow.test", "name", name),
					resource.TestCheckResourceAttr("n8n_workflow.test", "active", "false"),
					resource.TestCheckResourceAttrSet("n8n_workflow.test", "id"),
				),
			},
			{
				ResourceName:            "n8n_workflow.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"nodes", "connections", "settings"},
			},
			{
				Config: testAccWorkflowConfig(name+"-updated", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_workflow.test", "name", name+"-updated"),
					resource.TestCheckResourceAttr("n8n_workflow.test", "active", "true"),
				),
			},
		},
	})
}

func testAccWorkflowConfig(name string, active bool) string {
	return fmt.Sprintf(`
provider "n8n" {
  # endpoint and api_key from N8N_ENDPOINT / N8N_API_KEY
}

resource "n8n_workflow" "test" {
  name   = %q
  active = %t

  # webhook (not manualTrigger): n8n rejects activate without a trigger/webhook/polling node.
  nodes = jsonencode([
    {
      id          = "webhook"
      name        = "Webhook"
      type        = "n8n-nodes-base.webhook"
      typeVersion = 2
      position    = [0, 0]
      webhookId   = "tf-acc-%d"
      parameters = {
        path       = "tf-acc-%d"
        httpMethod = "GET"
      }
    }
  ])

  connections = jsonencode({})
  settings    = jsonencode({})
}
`, name, active, os.Getpid(), os.Getpid())
}
