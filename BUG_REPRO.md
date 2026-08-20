# Bug Reproduction

## Bug
The reviewed-to-published state transition is inconsistent across the domain transition table, repository CAS, service, and HTTP response.

## Trigger
Move a flag into review, approve it, and ask the publishing path to advance the same version.

## Error
The publishing tests reject a legal transition, and the observable state can fall back to an older value.
