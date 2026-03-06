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
