# Bug Reproduction

## Bug
Targeting previews, saved plans, and caller input share slice backing arrays.

## Trigger
Preview one plan, append another segment, save a second version, and then inspect the first version and original input.

## Error
The targeted tests show mutated segment IDs or members and occasionally a missing segment.
