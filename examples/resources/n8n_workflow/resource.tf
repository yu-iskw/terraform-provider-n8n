resource "n8n_workflow" "example" {
  name = "example-manual-trigger"
  # Manual-trigger workflows cannot be activated; use a webhook/schedule/polling
  # trigger (or keep active=false) when you need an active workflow.
  active = false

  nodes = jsonencode([
    {
      id          = "manual"
      name        = "When clicking 'Execute workflow'"
      type        = "n8n-nodes-base.manualTrigger"
      typeVersion = 1
      position    = [0, 0]
      parameters  = {}
    }
  ])

  connections = jsonencode({})
  settings    = jsonencode({})
}
