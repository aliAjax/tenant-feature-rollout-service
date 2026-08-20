# Bug Reproduction

## Bug
A default environment has a nil policy map, and typed-nil validators are treated as usable values.

## Trigger
Create an environment without policies, append the first protection rule, and invoke the validation path with a typed-nil validator.

## Error
The policy write panics with `assignment to entry in nil map`; the validator path also accepts an unusable value.
