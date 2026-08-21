# Testing setup reference

Utilities used by provider tests in `internal/provider`.

## `isIntegrationTestMode()`

- Defined in `internal/provider/utils.go`.
- Returns true when `TF_ACC` is `1`.
- Skips acceptance-style tests during normal `make test` / `go test` runs.

## `testAccPreCheck(t)`

- Defined in `internal/provider/provider_test.go`.
- Skips unless `TF_ACC=1`. Requires `N8N_ENDPOINT` and `N8N_API_KEY` for live-API acceptance tests.
- Prefer `make testacc-docker` (local Compose + bootstrap) over a hosted instance; see [run-acceptance-tests](../run-acceptance-tests/SKILL.md).
- Prefer `make testacc-docker` (local Compose + bootstrap) over a hosted instance; see [run-acceptance-tests](../run-acceptance-tests/SKILL.md).

## `testAccProtoV6ProviderFactories`

- Map key must match the provider local name in HCL: `"n8n"`.
- Uses `providerserver.NewProtocol6WithError(New("test")())`.

## Adding acceptance tests

- Build a `provider "n8n" { ... }` block (or rely on `N8N_ENDPOINT` / `N8N_API_KEY`).
- Example resource address: `n8n_workflow.example` (see `examples/resources/n8n_workflow/resource.tf`).
