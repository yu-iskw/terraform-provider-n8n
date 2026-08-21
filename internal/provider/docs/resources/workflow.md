Manages an n8n workflow via the Public API.

Provide `nodes`, `connections`, and optional `settings` as JSON strings. Activation is controlled with `active` after the workflow document is written. Activation requires a trigger, webhook, or polling node — a manual-trigger-only workflow cannot be activated.
