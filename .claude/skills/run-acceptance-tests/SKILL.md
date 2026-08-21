---
name: run-acceptance-tests
description: Guide execution of Terraform provider acceptance tests (TF_ACC) with correct environment setup and targeted runs.
---

# Run Acceptance Tests

## Purpose

Standardize running acceptance tests in this repository. Acceptance tests use `terraform-plugin-testing` with `TF_ACC=1` and call a live n8n Public API (`/api/v1` + `X-N8N-API-KEY`).

## Prerequisites

Live-API tests need `N8N_ENDPOINT` and `N8N_API_KEY` (see `testAccPreCheck` in `internal/provider/provider_test.go`). Prefer the Docker path below so you do not need a hosted instance.

### Option A: Local Docker (recommended)

Requires Docker Compose v2, `curl`, and `python3`.

- **Command**: `make testacc-docker`
- **Details**: Starts [`docker-compose.dev.yml`](../../../docker-compose.dev.yml), runs [`scripts/bootstrap-n8n.sh`](../../../scripts/bootstrap-n8n.sh) to mint an API key, then `TF_ACC=1 go test ./internal/provider/...`.
- Targeted: `make testacc-docker TESTARGS='-run ^TestAccN8nWorkflow_'`

### Option B: External instance

- Copy `.env.template` to `.env`: `cp .env.template .env`
- Set `N8N_ENDPOINT` and `N8N_API_KEY`, then `make testacc`.

## Workflow

### 1. Full acceptance suite (already-configured endpoint)

- **Command**: `make testacc`
- **Details**: Runs `go test ./internal/provider/...` with `TF_ACC=1`. Expects `N8N_ENDPOINT` / `N8N_API_KEY` already in the environment.

### 2. Targeted acceptance tests

- **Command**: `make testacc TESTARGS="-run <Pattern>"`
- **Example**: `make testacc TESTARGS="-run ^TestAccN8nWorkflow_"`

## Distinction from unit tests

- **Unit tests**: `make test`. Does not set `TF_ACC`.
- **Acceptance tests (Docker)**: `make testacc-docker`. Starts local n8n, then sets `TF_ACC=1`.
- **Acceptance tests (external)**: `make testacc`. Sets `TF_ACC=1` against a pre-existing endpoint.

## Troubleshooting

- **Timeouts**: `make testacc TESTARGS="-timeout 120m"` or `make testacc-docker TESTARGS="-timeout 20m"`.
- **Compose not healthy**: `docker compose -f docker-compose.dev.yml ps` and logs; healthcheck uses `/healthz/readiness` via Node inside the image.
- **Bootstrap CSRF/401**: ensure `browser-id` is sent (script does this) and `N8N_SECURE_COOKIE=false` for HTTP localhost.
- **Authentication**: Verify `N8N_ENDPOINT` / `N8N_API_KEY` match what `testAccPreCheck` expects.
- **Cleanup**: After failures against an external API, remove stray resources. Docker fixture is ephemeral (no volume); `docker compose -f docker-compose.dev.yml down -v` resets it.
