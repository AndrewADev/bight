# Developing

## Prerequisites

- Go (see `go.mod` for minimum version)
- [prek](https://github.com/j178/prek) — a faster drop-in for `pre-commit` (`pre-commit` also works if preferred)

## Setup

```bash
prek install   # or: pre-commit install
```

## Pre-commit hooks

This repo uses [`prek`](https://github.com/j178/prek) to run pre-commit hooks defined in `.pre-commit-config.yaml`. The config is fully compatible with [`pre-commit`](https://pre-commit.com) if you prefer that instead.

Hooks run automatically on `git commit`, or manually with:

```bash
prek run --all-files
```

## Commit messages

Commit messages follow [Angular commit conventions](https://github.com/angular/angular/blob/main/CONTRIBUTING.md#commit). Lint a message with:

```bash
conform enforce --commit-msg-file .git/COMMIT_EDITMSG
```

## Releasing

Releases are triggered by pushing a semver tag. The CI workflow builds 5-platform binaries, generates a changelog via `git-cliff`, and publishes a GitHub release.

Prerequisites: [`git-cliff`](https://git-cliff.org/docs/installation).

```bash
# 1. Update CHANGELOG.md and stage the release commit
git-cliff --unreleased --tag v0.2.0 --prepend CHANGELOG.md
git commit -am "chore: release v0.2.0"

# 2. Tag and push
git tag v0.2.0
git push && git push --tags
```

CI will publish the release automatically once the tag is pushed.

### Versioning

Format: `v0.MINOR.PATCH` (major is pinned at `0` until the config schema and CLI are stable).

- **Minor bump** (`v0.2.0 → v0.3.0`): may include breaking changes to `.bight.yml` or the CLI interface.
- **Patch bump** (`v0.2.0 → v0.2.1`): bug fixes only, no breaking changes.
