# Terraform Provider for n8n

Manage [n8n](https://n8n.io/) workflows as code with Terraform, using the [n8n Public API](https://docs.n8n.io/connect/n8n-api/).

## Example Usage

```hcl
terraform {
  required_providers {
    n8n = {
      source = "yu-iskw/n8n"
    }
  }
}

provider "n8n" {
  endpoint = "https://n8n.example.com"
  api_key  = var.api_key

  # Or export N8N_ENDPOINT and N8N_API_KEY instead.
  # max_concurrent_requests = 5
  # requests_per_second     = 20
}

resource "n8n_workflow" "example" {
  name   = "example-manual-trigger"
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

data "n8n_workflow" "example" {
  id = n8n_workflow.example.id
}
```

## Authentication

Create an API key in n8n under **Settings → n8n API**. The provider sends it as the `X-N8N-API-KEY` header.

## Development

Optional: with [mise](https://mise.jdx.dev/), run `mise trust` if prompted, then `mise install` to get the pinned Trunk CLI from `mise.toml` (Go is not installed via mise).

```shell
make test
go build -v ./
go generate ./...
```

Use `internal/provider` for Terraform wiring and `internal/n8n` for the Public API client.
