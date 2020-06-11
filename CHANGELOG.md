# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog], and this project adheres to
[Semantic Versioning].

<!-- references -->
[Keep a Changelog]: https://keepachangelog.com/en/1.0.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html

## [Unreleased]

## Added

- **[BC]** Add `assert.Assertion.End()`

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
[Unreleased]: https://github.com/dogmatiq/testkit
[0.1.0]: https://github.com/dogmatiq/testkit/releases/tag/v0.1.0
[0.1.1]: https://github.com/dogmatiq/testkit/releases/tag/v0.1.1
[0.2.0]: https://github.com/dogmatiq/testkit/releases/tag/v0.2.0
[0.3.0]: https://github.com/dogmatiq/testkit/releases/tag/v0.3.0
[0.4.0]: https://github.com/dogmatiq/testkit/releases/tag/v0.4.0

<!-- version template
## [0.0.1] - YYYY-MM-DD

### Added
### Changed
### Deprecated
### Removed
### Fixed
### Security
-->
