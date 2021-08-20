# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Changed
* Updated vault api lib to v1.1.1 to try to resolve dependabot resolution
  issues.

### Security
* Updated argo-workflows to v3.1.8 to address CVE-2021-37914
  (https://github.com/argoproj/argo-workflows/security/advisories/GHSA-h563-xh25-x54q).

## [0.8.1] - 2021-08-20
### Fixed
* Target credential_type only supports 'assumed_role'.

## [0.8.0] - 2021-08-20
### Changed
* Refactored validations.

## [0.7.0] - 2021-08-20
### Changed
* config.listFrameworks() are now sorted (helps avoid flaky tests).
* Transaction ID label added for workflow submissions.

### Fixed
* config.listFrameworks() was returning additional empty items.
* Invalid json response when creating a project with invalid admin credentials.
* Request logger didn't log all key/value pairs.
* Vault token exchange logic for examples when there's a failure.

## [0.6.3] - 2021-08-03
### Fixed
* Refactor credential provider client.
* Fix response statuses.

## [0.6.2] - 2021-07-29
### Fixed
* Parsing of csv arguments in cli with `=` in the value.

## [0.6.1] - 2021-07-27
### Fixed
* Fixed get project response shape.

## [0.6.0] - 2021-07-26
### Added
* `policy_document` to targets.

### Changed
* Updated go to 1.16.6.

## [0.5.1] - 2021-07-23
### Fixed
* Validations for workflow should not fail if arguments are empty

## [0.5.0] - 2021-07-23
### Added
* Git auth can now be configured for ssh or https

## [0.4.6] - 2021-07-19
### Changed
* Terraform example code no longer needs state bucket

## [0.4.5] - 2021-07-19
### Fixed
* Fixed a bug in create git workflow validation

## [0.4.4] - 2021-07-16
### Changed
* Refactor validations.
* Refactor requests and responses.

## [0.4.3] - 2021-07-01
### Fixed
* Fixed a bug while fetching existing repositories

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
