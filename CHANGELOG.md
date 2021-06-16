# Changelog

## [Unreleased]
### Added
* Release automation using GoReleaser.

## v0.3.0 (2021-06-16)
### Added
* Health check tests.
* More linting.

### Fixed
* Issues reported by linter.

### Changed
* Health check uses newer env var pattern.
* Health check uses newer logging pattern.
* Health check now responds with the same for success/failure (was json for
  failures).

## v0.2.1 (2021-06-16)
### Added
* CLI tests.
* More linting.

### Changed
* Break up CLI commands into separate files.

### Fixed
* Issues reported by linter.

## v0.2.0 (2021-06-15)
### Changed
* Move env to internal service package

## v0.1.3 (2021-06-14)
### Changed
* Using X-B3-TraceId as trace HTTP header

## v0.1.2 (2021-06-14)
### Changed
* Adding HTTP headers to Vault client for logging (e.g. transaction ID)

## v0.1.1 (2021-06-09)
### Fixed
* Add additional valid status codes for Vault health check

## v0.1.0 (2021-06-08)
### Added
* Tests for vault credential provider
* Vault service health check

### Changed
* Update credentials provider to be internal package
* Environmental variable handling

## v0.0.4 (2021-06-03)
### Fixed
* Passing Argo context to Argo Workflow calls

## v0.0.3 (2021-05-27)
### Changed
* Update environmental variable name to specify workflow execution namespace

### Added
* Adds operations API for git based sync/diff

### Fixed
* Fix typos in CLI command description
* Use GitHub Action for linting.
* Fix issues reported by linter.
* Perform deeper linting.
* Add build caching.
* Remove vendoring.

## v0.0.2 (2021-05-06)
### Changed
* Set service port via environment variable, default 8443

## v0.0.1 (2021-05-06)
### Added
* Initial release
