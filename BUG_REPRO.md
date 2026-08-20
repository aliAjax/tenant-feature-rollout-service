# Bug Reproduction

## Bug
Request cancellation is replaced or ignored between the HTTP adapter, batch service, repository, and downstream transport.

## Trigger
Run a batch evaluation with a short deadline while one downstream call is deliberately blocked.

## Error
The request expires, but background evaluation continues and delays the next batch.
