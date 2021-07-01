# Releasing

This project uses [GoReleaser](https://goreleaser.com/) and GitHub Actions for
releases. You can find the release workflow at
`./github/workflows/release.yaml`. The release workflow only triggers on tags
which start with `v` (e.g. `v0.1.0`).

To create a new release:

* Ensure the `CHANGELOG.md` is up to date (bump to the desired version, etc) on
  `main`.
* Locally, tag `main` with the desired semver version and push the tag up.

```shell
git tag v0.1.0 && git push --tags
```

The release process should begin. This will build/package artifacts and create
a GitHub release with the relevant `CHANGELOG.md` entries and artifacts
attached. If the release process does not find entries for the associated tag,
it will pull the entries in the `## [Unreleased]` section.

You can create a tag for a semver pre-release version, such as: `v0.1.0`
followed by a hyphen and characters (`v0.1.0-dev1`, `v0.1.0-alpha`, etc.). This
will result in creating a GitHub pre-release. This can be useful for testing
without releasing a "final" version.
