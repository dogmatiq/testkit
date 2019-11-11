# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog], and this project adheres to
[Semantic Versioning].

<!-- references -->
[Keep a Changelog]: https://keepachangelog.com/en/1.0.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html

## [Unreleased]

### Changed

- Test logs and assertion reports now display the file and line number of the test operation ([#38])

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

- Add `StartTime()` test option to configure the initial time of the test clock ([#21])

## [0.1.0] - 2019-08-01

- Initial release

<!-- references -->
[Unreleased]: https://github.com/dogmatiq/testkit
[0.1.0]: https://github.com/dogmatiq/testkit/releases/tag/v0.1.0
[0.1.1]: https://github.com/dogmatiq/testkit/releases/tag/v0.1.1
[0.2.0]: https://github.com/dogmatiq/testkit/releases/tag/v0.2.0
[0.3.0]: https://github.com/dogmatiq/testkit/releases/tag/v0.3.0

[#21]: https://github.com/dogmatiq/testkit/issues/21
[#38]: https://github.com/dogmatiq/testkit/issues/38

<!-- version template
## [0.0.1] - YYYY-MM-DD

### Added
### Changed
### Deprecated
### Removed
### Fixed
### Security
-->
