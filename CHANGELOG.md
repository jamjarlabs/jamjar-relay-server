# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 2021-14-18
### Added
- New `RESPONSE_CLIENT_DISCONNECT` message sent to the host when a client disconnects.

### Changed
- Host ID now sent along with `RESPONSE_FINISH_HOST_MIGRATE`.

## [0.3.0] - 2021-14-16
### Changed
- API responses now exposed under specs, rather than internal packages.
- Added `RESPONSE_CLIENT_CONNECT` message sent to a host when a new client connects.

## [0.2.0] - 2021-03-24
### Added
- Publishing protobuf specification files.

## [0.1.0] - 2021-03-23
### Added
- Initial release of jamjar-relay-server

[Unreleased]: https://github.com/jamjarlabs/jamjar-relay-server/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/jamjarlabs/jamjar-relay-server/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/jamjarlabs/jamjar-relay-server/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/jamjarlabs/jamjar-relay-server/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/jamjarlabs/jamjar-relay-server/releases/tag/v0.1.0
