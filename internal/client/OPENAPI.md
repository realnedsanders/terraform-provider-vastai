# OpenAPI client architecture

The provider vendors Vast.ai's official OpenAPI document from [`vast-ai/docs`](https://github.com/vast-ai/docs/blob/main/api-reference/openapi.yaml) in [`openapi/openapi.yaml`](openapi/openapi.yaml). The generated package owns documented HTTP methods, paths, and query encoding. The surrounding `internal/client` package remains the stable domain adapter used by Terraform resources.

## Generate

```bash
make generate-openapi
```

Generation uses the pinned `oapi-codegen` version in `openapi/generate.go` and the operation allowlist in `openapi/oapi-codegen.yaml`. The allowlist contains only operations used by this provider. The complete official specification is retained as input so upstream changes remain reviewable.

The adapter deliberately preserves behavior not supplied by generated code:

- canonical trailing slashes, which avoid 301 redirects that can change mutation methods to GET;
- live API compatibility overrides for endpoint deletion, volume deletion, and team invites where the published request shape differs from the official Python SDK;
- retry and rate-limit backoff;
- Bearer and compatibility query authentication;
- provider User-Agent and Terraform logging;
- `APIError` and HTTP 200 `success:false` handling;
- handwritten domain types and unusual response decoding;
- waiters, create-then-read logic, and client-side filtering.

## Handwritten fallbacks

The following provider calls remain on the raw transport because the official specification does not contain the route or contains a different, incompatible route:

- audit logs: `GET /api/v0/audit_logs`;
- clusters and members: `/api/v0/cluster`, `/api/v0/clusters`, and `/api/v0/cluster/remove_machine`;
- instance list: provider v0 `/api/v0/instances` versus documented v1 `/api/v1/instances`;
- instance template update: `PUT /api/v0/instances/update_template/{id}`;
- network volume create/search: `/api/v0/network_volumes` and `/api/v0/network_volumes/search`;
- overlays and members: `/api/v0/overlay`;
- volume copy: `POST /api/v0/volumes/copy`;
- worker groups: provider `/api/v0/autojobs` versus documented `/api/v0/workergroups`.

Do not switch these calls to a similarly named generated operation without live API verification. Remove a fallback when Vast.ai publishes the exact contract and its adapter tests pass.
