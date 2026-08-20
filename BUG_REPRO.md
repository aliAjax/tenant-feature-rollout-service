# Bug Reproduction

## Bug
Quota transition rules, repository CAS, compensation, and HTTP state reporting disagree after a failed commit.

## Trigger
Reserve quota, fail the commit, release or compensate it, and then attempt a new reservation.

## Error
The second reservation still reports occupied quota and the status endpoint exposes a stale state.
