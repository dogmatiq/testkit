# Migration Guide

## 0.11.0

This release is a major reworking of the entire testing API with the intent of
formalizing the idea of a test "action" as a distinct concept from an
"expectation" (previously known as an assertion).

If you are migrating tests from testkit v0.10.0 or prior, you will need to make
the following changes:

> **NOTE TO MAINTAINERS:** For the time being the new `Expectation` interface is just
a type alias for the existing `assert.Assertion` interface. Before this release
is tagged the `assert` package will be rolled into the root `testkit` package
and these instructions will be updated to match.

### `Test.ExecuteCommand()`, `RecordEvent()`, `AdvanceTime()` and `Call()`

These methods have been replaced by functions in the `testkit` package of the
same name. Each of these functions returns an `Action` which can be used with
the new `test.Expect()` method.

Individual actions can no longer accept a list of `engine.OperationOption`.
These must be set when the test is first begun.

**Before**
```go
test.ExecuteCommand(
    command,
    assert.EventRecorded(event),
)
```

**After**
```
test.Expect(
    testkit.ExecuteCommand(command),
    assert.EventRecorded(event),
)
```

### `Test.Prepare()`

This method now accepts a list of `Action` values instead `dogma.Message`. This
means you now need to nominate whether the message is a command or an event.

**Before**
```go
test.Prepare(command, event)
```

**After**
```go
test.Prepare(
    testkit.ExecuteCommand(command),
    testkit.RecordEvent(command),
)
```

### `assert.User`

- TODO. See note at top of this release.

### `assert.Nothing`

This assertion has been removed. It is no longer necessary as any of the
`Action` types can be used with `Prepare()` which does not make any assertions.

**Before**
```go
test.AdvanceTime(
    testkit.ByDuration(1 * time.Second),
    assert.Nothing,
)
```

**After**
```go
test.Prepare(
    testkit.AdvanceTime(testkit.ByDuration(1*time.Second)),
)
```
