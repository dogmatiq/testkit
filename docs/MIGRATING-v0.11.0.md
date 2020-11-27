# Migrating to Testkit v0.11.0

## Overview

This release includes extensive changes to the testing API. It formalizes the
concepts of "actions" and "expectations" with the goal of a more consistent API
and a greater level of extensibility.

### Actions

An action is an operation performed within a test that causes the Dogma
application being tested to do something. They are represented by the `Action`
interface.

The following functions in the `testkit` package each return an `Action`:

- `ExecuteCommand()`
- `RecordEvent()`
- `AdvanceTime()`
- `Call()`

Prior to this release the `Test` type had methods with these names. These
methods have been removed. Instead, the `Test.Expect()` method is used to
perform a single action and fails the test unless it meets a specific
expectation.

Actions can be performed without any expectations using `Test.Prepare()`.

### Expectations

An expectation is some criteria that an action is expected to meet. They are
represented by the `Expectation` interface.

The following functions in the `testkit` package each return an `Expectation`:

- `ToExecuteCommand()`
- `ToRecordEvent()`
- `ToExecuteCommandOfType()`
- `ToRecordEventOfType()`
- `AllOf()`
- `AnyOf()`
- `NoneOf()`
- `ToSatisfy()`

Prior to this release expectations were called "assertions". These functions
(some with slightly different names) were in the `assert` package, which has
been removed.

## Migrating existing tests

If you are migrating tests from testkit v0.10.0 or prior you will need to
accomodate the changes listed in this section.

Example code is provided showing how each feature was used both before and after
this release. Within these examples the `test` identifier refers to a
`testkit.Test` variable. It is assumed that the `testkit` package has been "dot
imported" meaning that any unqualified function call refers to a function in the
`testkit` package.

### `Runner` and `New()` have been removed

Prior to this release a `Test` was constructed by first constructing a `Runner`
then calling its `Begin()` or `BeginContext()` method. The `Runner` type has
been removed and its methods have been moved to functions in the `testkit`
package.

<table width="100%">
<thead><tr><td>Before</td><td>After</td></tr></head>
<tr valign="top"><td>

<!-- BEFORE -->
```go
test := New().Begin(t)
```

</td><td>

<!-- AFTER -->
```go
test := Begin(t)
```

</td></tr>
<tr><td></td><td></td></tr>
<tr valign="top"><td>

<!-- BEFORE -->
```go
test := New().BeginContext(ctx, t)
```

</td><td>

<!-- AFTER -->
```go
test := BeginContext(ctx, t)
```

</td></tr>
</table>

### `Test.Prepare()` now takes actions as parameters

Prior to this release `Prepare()` accepted `dogma.Message` parameters. It now
requires `Action` values instead.

This change allows any action to be performed without an expectation. This is
particularly useful with `AdvanceTime()` which was often used with the special
`assert.Nothing` assertion to advance the test's virtual clock without making
any real assertion.

<table width="100%">
<thead><tr><td >Before</td><td>After</td></tr></head>
<tr valign="top"><td>

<!-- BEFORE -->
```go
test.Prepare(
    SomeCommand{ /* ... */ },
    SomeEvent{ /* ... */ },
)
```

</td><td>

<!-- AFTER -->
```go
test.Prepare(
    ExecuteCommand(SomeCommand{ /* ... */ }),
    RecordEvent(SomeEvent{ /* ... */ }),
)
```

</td></tr>
<tr><td></td><td></td></tr>
<tr valign="top"><td>

<!-- BEFORE -->
```go
test.AdvanceTime(
    ByDuration(10 * time.Second),
    assert.Nothing,
)
```

</td><td>

<!-- AFTER -->
```go
test.Prepare(
    AdvanceTime(ByDuration(10 * time.Second)),
)
```

</td></tr>
</table>

### `Test.ExecuteCommand()`, `RecordEvent()`, `AdvanceTime()` and `Call()` have been removed

These methods on `Test` have been replaced with functions of the same name, each
of which returns an `Action`.

This change allows these actions to be performed without an expectation using
`Test.Prepare()`.

<table width="100%">
<thead><tr><td >Before</td><td>After</td></tr></head>
<tr valign="top"><td>

<!-- BEFORE -->
```go
test.ExecuteCommand(
    SomeCommand{ /* ... */ },
    assert.EventRecorded(SomeEvent{ /* ... */ }),
)
```

</td><td>

<!-- AFTER -->
```go
test.Expect(
    ExecuteCommand(SomeCommand{ /* ... */ }),
    ToRecordEvent(SomeEvent{ /* ... */ }),
)
```

</td></tr>
<tr><td></td><td></td></tr>
<tr valign="top"><td>

<!-- BEFORE -->
```go
test.RecordEvent(
    SomeEvent{ /* ... */ },
    assert.CommandExecuted(SomeCommand{ /* ... */ }),
)
```

</td><td>

<!-- AFTER -->
```go
test.Expect(
    RecordEvent(SomeEvent{ /* ... */ }),
    ToExecuteCommand(SomeCommand{ /* ... */ }),
)
```

</td></tr>
<tr><td></td><td></td></tr>
<tr valign="top"><td>

<!-- BEFORE -->
```go
test.AdvanceTime(
    ByDuration(10 * time.Second),
    assert.EventRecorded(SomeEvent{ /* ... */ }),
)
```

</td><td>

<!-- AFTER -->
```go
test.Expect(
    AdvanceTime(ByDuration(10 * time.Second)),
    ToRecordEvent(SomeEvent{ /* ... */ }),
)
```

</td></tr>
<tr><td></td><td></td></tr>
<tr valign="top"><td>

<!-- BEFORE -->
```go
test.Call(
    func() error { return nil },
    assert.EventRecorded(SomeEvent{ /* ... */ }),
)
```

</td><td>

<!-- AFTER -->
```go
test.Expect(
    Call(func() { /* no error is returned ¹ */ }),
    ToRecordEvent(SomeEvent{ /* ... */ }),
)
```

</td></tr>
</table>

### The `assert` package has been removed

The commonly used assertions that were in the `assert` package have been
reimplemented in `testkit` as expectations. Some of the function names have been
changed.

The `assert.Nothing` assertion has been removed. It is no longer necessary as
any action can be performed without an expectation using `Test.Prepare()`.

<table width="100%">
<thead><tr><td >Before</td><td>After</td></tr></head>
<tr valign="top"><td>

<!-- BEFORE -->
```go
assert.EventRecorded(SomeEvent{ /* ... */ })
```

</td><td>

<!-- AFTER -->
```go
ToRecordEvent(SomeEvent{ /* ... */ })
```

</td></tr>
<tr><td></td><td></td></tr>
<tr valign="top"><td>

<!-- BEFORE -->
```go
assert.EventTypeRecorded(SomeEvent{})
```

</td><td>

<!-- AFTER -->
```go
ToRecordEventOfType(SomeEvent{})
```

</td></tr>
<tr><td></td><td></td></tr>
<tr valign="top"><td>

<!-- BEFORE -->
```go
assert.CommandExecuted(SomeCommand{ /* ... */ })
```

</td><td>

<!-- AFTER -->
```go
ToExecuteCommand(SomeCommand{ /* ... */ })
```

</td></tr>
<tr><td></td><td></td></tr>
<tr valign="top"><td>

<!-- BEFORE -->
```go
assert.CommandTypeExecuted(SomeCommand{})
```

</td><td>

<!-- AFTER -->
```go
ToExecuteCommandOfType(SomeCommand{})
```

</td></tr>
<tr><td></td><td></td></tr>
<tr valign="top"><td>

<!-- BEFORE -->
```go
assert.AllOf(/* ... */)
```

</td><td>

<!-- AFTER -->
```go
AllOf(/* ... */)
```

</td></tr>
<tr><td></td><td></td></tr>
<tr valign="top"><td>

<!-- BEFORE -->
```go
assert.AnyOf(/* ... */)
```

</td><td>

<!-- AFTER -->
```go
AnyOf(/* ... */)
```

</td></tr>
<tr><td></td><td></td></tr>
<tr valign="top"><td>

<!-- BEFORE -->
```go
assert.NoneOf(/* ... */)
```

</td><td>

<!-- AFTER -->
```go
NoneOf(/* ... */)
```

</td></tr>
<tr><td></td><td></td></tr>
<tr valign="top"><td>

<!-- BEFORE -->
```go
assert.Should(
    "do something",
    func(s *assert.S) { /* ... */ },
)
```

</td><td>

<!-- AFTER -->
```go
ToSatisfy(
    "do something",
    func(t *SatisfyT) { /* ... */ },
)
```

</td></tr>
</table>

### Enabling and disabling message handlers

Within a `Test` both projection and integration message handler types are
disabled by default. Prior to this release such handlers could be enabled using
`WithOperationOptions(...)`.

As of this release enabling or disabling handlers by type is discouraged
<sup>2</sup>. Instead, individual handlers are enabled and disabled by name
using the `Test.EnableHandlers()` and `DisableHandlers()` methods.

This change is made to accomodate future changes to the expectation and
reporting systems that will analyse each handler's routing configuration to
eliminate impossible expectations and provide more meaningful failure reports.

<table width="100%">
<thead><tr><td>Before</td><td>After</td></tr></head>
<tr valign="top"><td>

<!-- BEFORE -->
```go
test := New().Begin(
    t,
    WithOperationOptions(
        engine.EnableProjections(true),
    ),
)
```

</td><td>

<!-- AFTER -->
```go
test := Begin(t).
    EnableHandlers("some-projection")
```

</td></tr>
</table>

---

<strong><sup>1</sup></strong> The function passed to `Call()` no longer
returns an `error`. Use the standard features of your testing framework to
ensure no errors occur within the function.

<strong><sup>2</sup></strong> For the time being it is still possible to set
engine operation options within a `Test` using `WithUnsafeOperationOptions()`.
This approach provides no guarantees as to how these options will interact with
the operation options that are set automatically by the `Test`.
