# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog], and this project adheres to
[Semantic Versioning].

<!-- references -->
[Keep a Changelog]: https://keepachangelog.com/en/1.0.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html

## [Unreleased]

This release is a major reworking of the entire testing API with the intent of
formalizing the idea of a test "action" as a distinct concept from an
"expectation" (previously known as an assertion).

Please see the [MIGRATING.md] file for more detailed instructions on how to
migrate tests from prior versions.

### Added

- Add `Begin()` and `BeginContext()`
- Add `Action` interface and associated `ActionScope` type
- Add `Expectation`
- Add `Test.Expect()` method and associated `ExpectOption` and `ExpectOptionSet` types
- Add `testkit.ExecuteCommand()`, `RecordEvent()`, `AdvanceTime()` and `Call()`
- Add `TimeAdjuster` interface, for use with `AdvanceTime()`
- Add `engine.EnableHandler()`
- Add `Test.EnableHandlers()` and `DisableHandlers()`
- **[BC]** Add `TestingT.Failed()`, `Fatal()` and `Helper()` methods

### Changed

- **[BC]** Change `Test.Prepare()` to accept `...Action` (previously `...dogma.Message`)
- **[BC]** The function passed to `Call()` no longer returns an `error`
- **[BC]** `engine.New()` and `MustNew()` now accept `configkit.RichApplication` (previously `dogma.Application`)
- **[BC]** Rename `WithStartTime()` to `StartTimeAt()`
- **[BC]** Rename `WithOperationOptions()` to `WithUnsafeOperationOptions()`

### Removed

- **[BC]** Remove `New()` (use `Begin[Context]()` instead)
- **[BC]** Remove `Runner`
- **[BC]** Remove `Test.ExecuteCommand()`, `RecordEvent()`, `AdvanceTime()` and `Call()` methods
- **[BC]** Remove `TimeAdvancer` function (replaced with `TimeAdjustment` interface)
- **[BC]** Remove `WithEngineOptions()`

## [0.10.0] - 2020-11-16

### Changed

- **[BC]** Updated to Dogma to v0.10.0

### Added

- Add support for projection compaction

### Fixed

- Fix issue where engine operation options were not taking precedence over test options

## [0.9.0] - 2020-11-06

### Changed

- **[BC]** Updated to Dogma to v0.9.0

## [0.8.1] - 2020-11-04

### Added

- Add `engine.MustNew()`

## [0.8.0] - 2020-11-03

### Changed

- **[BC]** Remove aggregate `Create()` and `Exists()` methods (complies with Dogma v0.8.0)

## [0.7.0] - 2020-11-02

### Added

- Add `Test.AdvanceTime()`, `ByDuration()` and `ToTime()`
- Add `Test.Call()`
- Add `assert.Nothing`
- Add `engine.CommandExecutor`
- Add `engine.EventRecorder`
- **[BC]** Add `op` parameter to `assert.Assertion.Begin()`

### Removed

- **[BC]** Remove `Test.AdvanceTimeTo()` and `AdvanceTimeBy()`
- **[BC]** Remove `verbose` parameter `assert.Assertion.BuildReport()`
- **[BC]** Remove `RunnerVerbose()` and `Verbose()` options
- **[BC]** Remove `Engine.ExecuteCommand()` and `RecordEvent()`

### Changed

- The `-v` (verbose) option to `go test` no longer affects testkit's behavior, full logs are always rendered

## [0.6.2] - 2020-10-19

### Changed

- `Test.AdvanceTimeBy()` and `AdvanceTimeTo()` can now be called without making any assertions by passing a `nil` assertion

## [0.6.1] - 2020-10-19

## Added

- Add `Exists()` and `HasBegun()` methods to aggregate & process scopes

## [0.6.0] - 2020-06-12

## Changed

- **[BC]** Renamed `T` to `TestingT`, to avoid conflicts with `*testing.T` when dot-importing
- **[BC]** Renamed `assert.T` to `S`, to avoid conflicts with `*testing.T` when dot-importing

## [0.5.0] - 2020-06-11

## Added

- **[BC]** Add `assert.Assertion.End()`
- Add `engine.Run()` and `RunTimeScaled()`
- `engine.Engine` now protects its internal state with a mutex, allowing concurrent use

## Changed

- **[BC]** `assert.Should()` now provides an `assert.T` to the user instead of checking the `error` return value
- **[BC]** Rename `assert.Assertion.Prepare()` to `Begin()`
- **[BC]** Add `verbose` parameter to `assert.Assertion.BuildReport()`

## [0.4.0] - 2020-02-04

### Added

- Add `assert.Should()` for specifying user-defined assertions

### Changed

- **[BC]** Switched from `enginekit` to `configkit` for application configurations
- **[BC]** Flatten fields from `Envelope.Correlation` directly into `Envelope`
- Test logs and assertion reports now display the file and line number of the test operation

### Removed

- **[BC]** Removed the `assert.CompositeAssertion` type
- **[BC]** Removed the `assert.MessageAssertion` type
- **[BC]** Removed the `assert.MessageTypeAssertion` type

### Fixed

- `assert.CommandExecuted()` now renders the correct suggestions when it fails due to mismatched message role

## [0.3.0] - 2019-10-31

### Changed

- Updated EngineKit to v0.8.0

## [0.2.0] - 2019-10-21

### Changed

- **[BC]** Rename `StartTime()` option to `WithStartTime()` for consistency

## [0.1.1] - 2019-10-18

### Added

- Add `StartTime()` test option to configure the initial time of the test clock

## [0.1.0] - 2019-08-01

- Initial release

<!-- references -->
[MIGRATING.md]: https://github.com/dogmatiq/testkit/blob/main/MIGRATING.md
[Unreleased]: https://github.com/dogmatiq/testkit
[0.1.0]: https://github.com/dogmatiq/testkit/releases/tag/v0.1.0
[0.1.1]: https://github.com/dogmatiq/testkit/releases/tag/v0.1.1
[0.2.0]: https://github.com/dogmatiq/testkit/releases/tag/v0.2.0
[0.3.0]: https://github.com/dogmatiq/testkit/releases/tag/v0.3.0
[0.4.0]: https://github.com/dogmatiq/testkit/releases/tag/v0.4.0
[0.5.0]: https://github.com/dogmatiq/testkit/releases/tag/v0.5.0
[0.6.0]: https://github.com/dogmatiq/testkit/releases/tag/v0.6.0
[0.6.1]: https://github.com/dogmatiq/testkit/releases/tag/v0.6.1
[0.6.2]: https://github.com/dogmatiq/testkit/releases/tag/v0.6.2
[0.7.0]: https://github.com/dogmatiq/testkit/releases/tag/v0.7.0
[0.8.0]: https://github.com/dogmatiq/testkit/releases/tag/v0.8.0
[0.8.1]: https://github.com/dogmatiq/testkit/releases/tag/v0.8.1
[0.9.0]: https://github.com/dogmatiq/testkit/releases/tag/v0.9.0
[0.10.0]: https://github.com/dogmatiq/testkit/releases/tag/v0.10.0

<!-- version template
## [0.0.1] - YYYY-MM-DD

### Added
### Changed
### Deprecated
### Removed
### Fixed
### Security
-->
