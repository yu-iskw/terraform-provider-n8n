# Contributing

This repository is the Terraform provider for n8n (`yu-iskw/n8n`), built with the Terraform Plugin Framework. It manages n8n **team projects**, not workflows.

## Prerequisites

- Go, using the version declared in `go.mod` (install Go yourself; it is not provided via mise here)
- Terraform, using the version declared in `.terraform-version`
- GNU Make

Optional: [mise](https://mise.jdx.dev/) — run `mise trust` in the repo root if needed, then `mise install` to install the **Trunk CLI** only from [`mise.toml`](mise.toml). Go stays outside mise. When upgrading Trunk, update both `.trunk/trunk.yaml` `cli.version` and `mise.toml` `trunk = "..."`.

Quality checks use **Trunk** (`trunk check` / `trunk fmt`) plus **`gosec`** and **`deadcode`** via `GNUmakefile` targets (`make lint`, `make build`). Git hooks: run `make setup-dev` to clear conflicting `core.hooksPath` and run `trunk git-hooks sync` (no pre-commit framework).

## Local Development

Run the common checks:

```shell
make test
go build -v ./
go generate ./...
```

Run formatting before sending changes:

```shell
make format
```

## Repository Structure

- `main.go`: provider server entry point
- `internal/provider`: provider implementation, resources, data sources, docs embedded by tests, and unit tests
- `internal/api/controllers`: Terraform-facing orchestrators used by resources and data sources
- `internal/n8n`: n8n Public API HTTP client (`Client.HTTP`, `X-N8N-API-KEY`, optional rate or concurrency limits) plus `models/`, versioned `api/v1/<resource>/`, and `services/`
- `examples`: Terraform examples used by documentation generation
- `docs`: generated or hand-maintained provider documentation
- `tools`: Go tool dependency tracking for code generation and analysis
- `mise.toml`: optional Trunk CLI pin for mise users

## Acceptance tests

Team-project CRUD on a **licensed** n8n needs `feat:projectRole:admin`. Community Docker returns HTTP 403 for those APIs.

### Docker fixture (CRUD, import, delete protection)

`make testacc-docker` builds [`docker-compose.acc.yml`](docker-compose.acc.yml), a Public API stand-in that implements the verified GET/POST/PUT/DELETE `/api/v1/projects` contract. Use this path for local and CI acceptance tests:

```shell
make testacc-docker
make testacc-docker TESTARGS='-run ^TestAccN8nProject_'
```

### Licensed instance

Copy `.env.template` to `.env`, set `N8N_ENDPOINT` and `N8N_API_KEY` for a licensed n8n, then:

```shell
make testacc TESTARGS='-run ^TestAccN8nProject_'
```

### Community n8n (optional)

[`docker-compose.dev.yml`](docker-compose.dev.yml) still starts pinned Community n8n for live probes. Project acceptance tests skip there with a license 403.

`make test` stays unit-only (`TF_ACC` unset). Do not point acceptance tests at production n8n.
