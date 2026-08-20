# Bug Reproduction

## Bug
Duplicate project conflicts lose their sentinel identity across the domain, repository, service, and HTTP layers, so the endpoint returns an internal error instead of a conflict.

## Trigger
Register the same project name twice for one tenant, then inspect the returned HTTP status and `errors.Is` result at each layer.

## Error
The targeted tests report that `ErrProjectConflict` cannot be recognized and the HTTP response is not `409 Conflict`.
