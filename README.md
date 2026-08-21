# Terraform Provider for n8n

Manage [n8n](https://n8n.io/) **team projects** as code with Terraform, using the [n8n Public API](https://docs.n8n.io/connect/n8n-api/projects).

Team-project APIs require an n8n license that includes `feat:projectRole:admin` (Enterprise / a licensed instance). Community self-hosted n8n returns HTTP 403 for these operations.

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

resource "n8n_project" "platform" {
  name              = "platform"
  delete_protection = false
}

data "n8n_project" "platform" {
  id = n8n_project.platform.id
}

data "n8n_projects" "all" {}
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

Use `internal/provider` for Terraform wiring, `internal/api/controllers` for resource/data-source orchestration, and `internal/n8n` for the Public API client (`models`, versioned `api/v1/<resource>`, `services`).
