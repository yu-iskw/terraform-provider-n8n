# AGENTS.md

Shared guidance for agents working in this repository.

## Overview

Terraform provider for [n8n](https://n8n.io/), written in Go with the HashiCorp Terraform Plugin Framework. It talks to the n8n Public API and currently manages **team projects** (`n8n_project` resource, `n8n_project` / `n8n_projects` data sources). It does not manage workflows.

## mise (optional)

If you use [mise](https://mise.jdx.dev/), run `mise trust` in the repo root on first use if mise refuses to load the config, then `mise install`. [`mise.toml`](mise.toml) installs **only the Trunk CLI** (pinned to match `.trunk/trunk.yaml` `cli.version`). **Go is not installed via mise**; use your normal Go install and the `go` / `toolchain` lines in `go.mod`. When you upgrade Trunk, bump `cli.version` in `.trunk/trunk.yaml` and the `trunk` version in `mise.toml` together.

## Key Commands

- Unit tests: `make test`
- Build: `go build -v ./`
- Generate docs: `go generate ./...`
- Format: `make format`
- Lint: `make lint`

## Project Structure

- `main.go`: provider server entry point (`registry.terraform.io/yu-iskw/n8n`)
- `internal/provider`: provider schema, configuration, resources, data sources, embedded docs, and tests
- `internal/api/controllers`: Terraform-facing orchestrators (`n8n_project` resource and data sources call these, not services directly)
- `internal/n8n`: n8n Public API HTTP client (`X-N8N-API-KEY`, rate and concurrency limits), plus `models/`, versioned `api/v1/<resource>/` (one HTTP op per file), and `services/`
- `examples`: Terraform examples used by docs generation
- `docs`: provider documentation
- `tools`: Go tool dependencies
- `mise.toml`: optional Trunk CLI version for mise users (no Go via mise)

## Development Notes

- Prefer real, deterministic tests over mocks.
- Acceptance tests need `TF_ACC=1`, `N8N_ENDPOINT`, and `N8N_API_KEY`.
- `make testacc-docker` runs project acceptance tests (CRUD, import, delete protection) against a Public API fixture. Community n8n Docker still returns 403 for `feat:projectRole:admin`.
- Git hooks for this repo are **Trunk-only** (`make setup-dev` runs `trunk git-hooks sync`); there is no separate pre-commit install step.
