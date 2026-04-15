# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`bight` is a Go CLI tool that patches `.env` files on git branch checkout, keeping local environments in sync with the current branch. It hooks into git's `post-checkout` event and updates specific env vars based on configurable strategies.

## Commands

```bash
go build .              # build binary
go build ./...          # compile-check all packages (does not write binary)
go test ./...           # run all tests
go test ./... -run TestName  # run a single test
go vet ./...            # lint
prek install            # install pre-commit hooks (once, after cloning)
conform enforce --commit-msg-file .git/COMMIT_EDITMSG  # lint commit message (angular conventions)
```

## Architecture

The binary serves two roles: an **installer** (run by the user) and a **hook handler** (run by git).

**Entry points:**
- `bight install` — writes `.git/hooks/post-checkout` using `os.Executable()` to get the current binary path
- `bight post-checkout <prev> <new> <flag>` — called by git; reads branch info from the hook args (not via shell or go-git)

**Core flow on checkout:**
1. Load `.bight.yml` (repo-level), merging with `~/.bight.yml` (global defaults)
2. Determine which vars to patch based on the `on` trigger (`checkout`)
3. Apply strategy per var: `template` (branch/project interpolation), `random` (fresh value), `deterministic` (hashed/transformed from branch name — stable per branch)
4. Patch target `.env` files in-place using `godotenv` (not full replacement)

**Key dependencies:**
- `godotenv` — reading and writing `.env` files

## Code Style

- Prefer idiomatic Go: avoid shadowing built-in identifiers (`error`, `len`, `close`, etc.), use explicit names over clever ones, and follow standard Go naming conventions.
- Keep dependencies minimal — prefer zero-dependency solutions using the standard library where practical.

## Config

Per-repo config lives in `.bight.yml`. Global defaults in `~/.bight.yml`. The `on` trigger field controls when a var is patched: `checkout` (every branch switch).