# `internal/n8n`

HTTP client for the [n8n Public API](https://docs.n8n.io/connect/n8n-api/).

## Behavior

- **`Client.HTTP`**: standard `*http.Client` whose transport applies, in order: **max concurrent in-flight requests** (`max_concurrent_requests`), **token-bucket rate limit** (`requests_per_second`), then **`X-N8N-API-KEY`** using the provider `api_key`.
- **Endpoint**: instance URL is normalized to end with `/api/v1`.
- **Defaults** (when provider omits optional attributes): `10` concurrent requests, `10` requests per second.

Provider attributes are defined in [`internal/provider/provider.go`](../provider/provider.go).
