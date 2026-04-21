# Failing Predicate — Design Notes

## Context

This document describes an incomplete design problem discovered while resolving
GitHub issue #162 (guard checks that returned bare errors should instead
populate the test report).

The work is on the `fix/issue-162` branch, though recent commits may have been
squashed onto `main`. Check `git log --oneline` to confirm current state.

## What Was Done

### `failingPredicate` (expectation.go)

A new `failingPredicate` type was introduced. It implements `Predicate` and
always fails, reporting a pre-determined `criteria` and `explanation`. It is
used when an expectation is known to be impossible before any action is run —
for example, when the message type is not in the application's route set.

### `isExpectationOnImpossibleType` (expectation.messagecommon.go)

A helper function `isExpectationOnImpossibleType(s PredicateScope, t
message.Type) (string, bool)` returns an explanation string and `true` when an
expectation targeting type `t` can never be satisfied given the application's
configuration. The callers (`messageExpectation`, `messageTypeExpectation`,
`messageMatchExpectation`) now return a `&failingPredicate{...}` from
`Predicate()` instead of a live predicate when this check fires.

Previously these returned a bare `error` from `Predicate()`, which was caught
in `Test.Expect()` and passed to `t.testingT.Fatal(err)`. That meant the
failure bypassed the report renderer entirely.

### `Test.Expect()` — action error path (test.go, ~line 120)

When `doAction` returns an error inside `Expect()`, the code currently
constructs an inline `&Report{Criteria: act.Caption(), Explanation: err.Error()}`
and logs it before calling `FailNow()`.

This is marked `// REVIEW: this is a hack - making a report here.` and is
**not the right final state** (see problem below).

### `dispatchAction.Do()` still returns an error (action.dispatch.go)

The "unrecognized message type" check inside `dispatchAction.Do()` still
returns a bare `error`:

```go
if !s.App.RouteSet().HasMessageType(mt) {
    return inflect.Errorf(...)
}
```

This is marked `// REVIEW: this comment was removed but nothing changed?`.
The original TODO comment was removed on the assumption that the `test.go`
error-to-report conversion handled it. That assumption is partially correct but
the approach is wrong (see below).

## The Problem

The current `test.go` hack has two issues:

1. **Wrong criteria.** When `doAction` errors, `p.Report()` is never called.
   The report shown uses `act.Caption()` as the criteria instead of
   `e.Caption()`. The report is therefore framed as an action failure, not an
   expectation failure — which is semantically wrong.

2. **All errors treated equally.** The blanket conversion applies to every
   `doAction` error, including errors that represent programming mistakes in
   the test itself:
   - `advanceTimeAction`: "would reverse time" — a test bug, should be
     `Fatal`, not a report
   - `dispatchAction`: "unrecognized type" — a configuration problem, suitable
     for a report
   - User-defined actions: could be anything

## Proposed Solution

Revert the `test.go` action-error-to-report change. Restore `Fatal(err)` for
genuine unexpected action errors.

Move the "unrecognized type" check in `dispatchAction.Do()` upstream to a new
optional interface on `Action`:

```go
// actionValidator may be implemented by an Action to perform early validation
// against the application configuration before the action is executed.
//
// If Validate returns a non-empty explanation string, the action is considered
// impossible and Expect() will return a failing predicate rather than running
// the action.
type actionValidator interface {
    Validate(s PredicateScope) (explanation string)
}
```

`Test.Expect()` checks this interface after building the `PredicateScope` and
before calling `e.Predicate(s)`:

```go
if v, ok := act.(actionValidator); ok {
    if explanation := v.Validate(s); explanation != "" {
        p := &failingPredicate{
            criteria:    e.Caption(),
            explanation: explanation,
        }
        // log header + report + FailNow
        ...
        return t
    }
}
```

`dispatchAction` implements `actionValidator`, moving the route-set check out
of `Do()` and into `Validate()`. `Do()` no longer returns an error for this
case.

Note: `Prepare()` does not call `actionValidator` — it has no expectation to
attach the failure to, and the bare `Fatal(err)` from `Do()` is appropriate
there since unrecognized-type on `Prepare()` is already tested and working.

## File Locations

- `expectation.go` — `failingPredicate`, `Expectation`, `Predicate` interfaces
- `expectation.messagecommon.go` — `isExpectationOnImpossibleType`
- `action.go` — `Action` interface (add `actionValidator` here or alongside)
- `action.dispatch.go` — `dispatchAction`, move check to `Validate()`
- `test.go` — `Test.Expect()`, revert error path, add `actionValidator` check

## Tests to Update

- `action.dispatch.command_test.go` and `action.dispatch.event_test.go`:
  The "unrecognized message type" tests currently use `xtesting.ExpectContains`
  on a bare error log string. After the change, the failure will go through the
  report path (same as `isExpectationOnImpossibleType` does), so these tests
  need updating to match the indented `  |     ` report format — but only for
  the `Expect()` variant. The `Prepare()` variant keeps the bare error.

  Actually: `action.dispatch.command_test.go` uses `tc.Prepare(...)` for the
  unrecognized-type test. If `Prepare()` still calls `Fatal(err)`, those tests
  stay unchanged. Confirm which path each test exercises before editing.

- `test_test.go`: The `TestTest_Expect / it fails the test if the action
returns an error` test uses a `noopAction` with a synthetic error. After the
  revert this should go back to calling `Fatal(err)` directly and the test
  should already pass unchanged.

## State of the Branch

All tests pass (`make precommit` succeeds) with the current hacky
implementation. The REVIEW comments mark the two places that need rework:

- `test.go` ~line 134
- `action.dispatch.go` ~line 75
