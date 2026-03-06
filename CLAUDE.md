# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`bight` is a Go CLI tool that patches `.env` files on git branch checkout, keeping local environments in sync with the current branch. It hooks into git's `post-checkout` event and updates specific env vars based on configurable strategies.

## Commands

```bash
go build ./...          # build
go test ./...           # run all tests
go test ./... -run TestName  # run a single test
go vet ./...            # lint
```

## Architecture

The binary serves two roles: an **installer** (run by the user) and a **hook handler** (run by git).

**Entry points:**
- `bight install` — writes `.git/hooks/post-checkout` using `os.Executable()` to get the current binary path
- `bight post-checkout <prev> <new> <flag>` — called by git; reads branch info from the hook args (not via shell or go-git)

**Core flow on checkout:**
1. Load `.bight.yml` (repo-level), merging with `~/.bight.yml` (global defaults)
2. Determine which vars to patch based on the `on` trigger (`checkout` vs `db_create`)
3. Apply strategy per var: `template` (branch/project interpolation), `random` (fresh value), `deterministic` (hashed/transformed from branch name — stable per branch)
4. Optionally create a Postgres database if `database.auto_create: true` and the DB doesn't exist
5. Patch target `.env` files in-place using `godotenv` (not full replacement)

**Key dependencies:**
- `godotenv` — reading and writing `.env` files
- `pgx` or postgres driver — creating databases if configured

## Config

Per-repo config lives in `.bight.yml`. Global defaults in `~/.bight.yml`. The `on` trigger field controls when a var is patched: `checkout` (every branch switch) or `db_create` (only when the DB is newly created). See ONE-PAGER.md for the full config schema.
