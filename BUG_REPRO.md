# Bug Reproduction

## Bug
Snapshot export defers cleanup inside a loop, skips rollback, and lets commit or close errors overwrite the first business error.

## Trigger
Export a batch where the middle snapshot write fails and later items remain available for processing.

## Error
Open leases accumulate, rollback is absent, and the caller receives a cleanup or commit error instead of the original write failure.
