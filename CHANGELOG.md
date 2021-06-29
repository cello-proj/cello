# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.2] - 2021-06-29
### Changed
* Updated go to 1.16.5 for release process.

## [0.4.1] - 2021-06-29
### Changed
* Updated go to 1.16.5.

## [0.4.0] - 2021-06-28
### Added
* Git information is now stored in a postgreSQL database on project creation

### Changed
* Git information is now retrieved from db and used for workflow operations
* Project create API now takes in repository
* Operations API no longer takes in repository

## [0.3.2] - 2021-06-17
### Fixed
* Updated CHANGELOG to 'Keep a Changelog' format.

## [0.3.1] - 2021-06-17
### Added
* Release automation using GoReleaser.

### Changed
* Updated CHANGELOG to 'Keep a Changelog' format.

## [0.3.0] - 2021-06-16
### Added
* Health check tests.
* More linting.

### Changed
* Health check uses newer env var pattern.
* Health check uses newer logging pattern.
* Health check now responds with the same for success/failure (was json for
  failures).

### Fixed
* Issues reported by linter.

## [0.2.1] - 2021-06-16
### Added
* CLI tests.
* More linting.

### Changed
* Break up CLI commands into separate files.

### Fixed
* Issues reported by linter.

## [0.2.0] - 2021-06-15
### Changed
* Move env to internal service package

## [0.1.3] - 2021-06-14
### Changed
* Using X-B3-TraceId as trace HTTP header

## [0.1.2] - 2021-06-14
### Changed
* Adding HTTP headers to Vault client for logging (e.g. transaction ID

## [0.1.1] - 2021-06-09
### Fixed
* Add additional valid status codes for Vault health check

## [0.1.0] - 2021-06-08
### Added
* Tests for vault credential provider
* Vault service health check

### Changed
* Update credentials provider to be internal package
* Environmental variable handling

## [0.0.4] - 2021-06-03
### Fixed
* Passing Argo context to Argo Workflow calls

## [0.0.3] - 2021-05-27
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

## [0.0.2] - 2021-05-06
### Changed
* Set service port via environment variable, default 8443

## [0.0.1] - 2021-05-06
### Added
* Initial release
