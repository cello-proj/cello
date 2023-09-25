# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.18.0] - 2023-09-25
### Changed
* Added quotes to environment variables to sanitize & prevent errors

## [0.17.0] - 2023-08-15
### Added
* Flow request type all the way down to argo workflow calls

## [0.16.0] - 2022-12-21
### Added
* Added DB healthcheck
### Changed
* Non-existent workflows return 404 on GET workflow calls
* Fixed return type for health check
* Panics changed to clean exits in main.go

## [0.15.2] - 2022-10-12
### Changed
* Flush Argo Workflow log stream

## [0.15.1] - 2022-10-11
### Changed
* Remove Argo Workflow status call from log streaming

## [0.15.0] - 2022-09-06
### Changed
* Delete tokens idempotent

## [0.14.3] - 2022-08-15
### Changed
* List workflows returns an empty array instead of null for targets with no workflows

## [0.14.2] - 2022-07-28
### Changed
* Listing workflows no longer calls a status for each workflow

## [0.14.1] - 2022-07-14
### Added
* Expiration time for token

## [0.14.0] - 2022-07-08
### Added
* Added schema updates to create tokens table
* Added golang-migrate to manage schema
* Delete project token
* List project tokens
* Create project token

## [0.13.3] - 2022-06-16
### Changed
* Added github repository from db to get projects response

## [0.13.2] - 2022-06-09
### Changed
* Fixed quickstart issues with standing up database
* Added log statement to clarify issues with project creation

## [0.13.1] - 2022-05-24
### Changed
* The default config file has been renamed to cello.yaml (when the CONFIG env var is NOT specified)

## [0.13.0] - 2022-05-02
### Changed
* Default workflow renamed from argo-cloudops-single-step-vault-aws.yaml to cello-single-step-vault-aws.yaml
* Example manifests now point to cello-single-step-vault-aws workflow
* Images moved to from argocloudops to celloproj
* BREAKING cli binary renamed from argo-cloudops to cello
* Environment variables prefix changed from ARGO_CLOUDOPS to CELLO (backwards compatible)

## [0.12.1] - 2022-03-14
### Changed
* POTENTIALLY BREAKING Release tarballs use cello naming, potentially breaking change for automation scripts

## [0.12.0] - 2022-03-10
### Changed
* Reference renames all around to support new cello-proj/cello repository location.

## [0.11.0] - 2022-03-01
### Added
* Add target update api operation

### Changed
* Change get target api response

### Fixed
* Validate project exists when listing targets

## [0.10.0] - 2022-01-26
* Add `exec` type/command

## [0.9.0] - 2021-12-14
### Added
* Approve list for pre and execute image URIs.

## [0.8.5] - 2021-10-22
### Added
* Retry logging on stream INTERNAL_ERROR errors

## [0.8.4] - 2021-10-21
### Added
* Add `X-Accel-Buffering=no` header to service response calls for log streaming.

## [0.8.3] - 2021-10-05
### Changed
* Bump github.com/argoproj/argo-workflows/v3 from 3.1.8 to 3.1.13.
* Bump github.com/aws/aws-sdk-go from 1.40.51 to 1.40.52.
* Bump github.com/go-git/go-git/v5 from 5.3.0 to 5.4.2.
* Bump github.com/google/uuid from 1.2.0 to 1.3.0.
* Bump github.com/go-kit/log from 0.10.0 to 0.12.0.
* Bump github.com/upper/db/v4 from 4.1.0 to 4.2.1.
* Bump google.golang.org/grpc from 1.40.0 to 1.41.0.
* Migrated github.com/go-kit/kit to github.com/go-kit/log.
* Update to go 1.17.1.

### Fixed
* Nil pointer error on bad auth header.
* Revert http listen and serve tls.

## [0.8.2] - 2021-09-28
### Added
* Example manifest files.

### Changed
* Bump github.com/aws/aws-sdk-go from 1.33.16 to 1.40.28.
* Bump github.com/google/go-cmp from 0.5.2 to 0.5.6.
* Bump github.com/spf13/cobra from 1.1.3 to 1.2.1.
* Updated vault api lib to v1.1.1 to try to resolve dependabot resolution
  issues.
* Change build test exceptions declaration (linting).
* Explicitly set http1.1 for service.
* Remove health check debug log messages.
* Retry logging on stream INTERNAL_ERROR errors.

### Security
* Updated argo-workflows to v3.1.8 to address CVE-2021-37914
  (https://github.com/argoproj/argo-workflows/security/advisories/GHSA-h563-xh25-x54q).

### Fixed
* Duplicated log entries when streaming logs.

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
