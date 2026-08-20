# Bug Reproduction

## Bug
Rollout worker accounting and result-channel ownership are split across goroutines and error branches.

## Trigger
Run three workers where one fails quickly and two complete later, with the race detector enabled.

## Error
The stream ends before all results arrive and can panic with `close of closed channel` or `send on closed channel`.
