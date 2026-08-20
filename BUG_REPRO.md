# Bug Reproduction

## Bug
Governance denial identity is flattened while domain, audit, service, and HTTP layers add context.

## Trigger
Evaluate a production rollout without approval credentials while the audit append path also returns a warning.

## Error
`errors.Is` cannot recognize the denial, so the endpoint reports an internal error instead of `403 Forbidden`.
