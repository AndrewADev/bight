# Developing

## Prerequisites

- Go (see `go.mod` for minimum version)
- [prek](https://github.com/j178/prek) — a faster drop-in for `pre-commit` (`pre-commit` also works if preferred)
- [task](https://taskfile.dev/installation/) — task runner used for release workflows

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

## Command docs

Per-command reference under `docs/commands/` is generated from the cobra command tree. Regenerate after changing the public-facing interface:

```bash
task docs   # or: go generate ./...
```

The generator lives at `tools/docgen/main.go`.

## Preview builds

The `Preview` workflow builds per-platform binaries for an open PR and uploads them as artifacts, so reviewers can try the changes without a Go toolchain. It can be dispatched from Actions → Preview → Run workflow, or via the helper task:

```bash
task preview-build -- 42   # dispatch the Preview workflow for PR #42
```

Requires `gh` CLI authenticated with write access to the repo.

Once the run completes, anyone with read access can download the artifact for their platform with `task preview-fetch -- 42` (see the README).

## Releasing

Releases are triggered by pushing a semver tag. The CI workflow builds 5-platform binaries and publishes a GitHub release. The release notes are the matching version section extracted from `CHANGELOG.md` (the single source of truth), so the changelog must be committed before tagging.

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
