# Bug Reproduction

## Bug
Distribution snapshots expose internal references while concurrent writers mutate the same tenant bundle.

## Trigger
Read and encode a published snapshot while another goroutine appends a distribution change to the same tenant.

## Error
Old snapshot contents change and `go test -race` reports overlapping reads and writes.
