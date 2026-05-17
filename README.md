# bight

Patches `.env` files automatically on `git checkout` — keeping local environments in sync with the current branch.

```sh
git checkout feat-login
# bight: .env → DB_NAME=myapp_feat-login
# bight: .env → JWT_SECRET=***
```

Only the listed vars are touched. The rest of your `.env` is left untouched.

## Installation

**With Homebrew:**

```sh
brew install andrewadev/tap/bight
```

**With Go:**

```sh
go install github.com/AndrewADev/bight@latest
```

**From a release binary:**

Download the binary for your platform from the [releases page](https://github.com/AndrewADev/bight/releases/latest):

| Platform | File |
|---|---|
| macOS (Apple Silicon) | `bight-darwin-arm64` |
| macOS (Intel) | `bight-darwin-amd64` |
| Linux (x86-64) | `bight-linux-amd64` |
| Linux (ARM64) | `bight-linux-arm64` |
| Windows | `bight-windows-amd64.exe` |

Then make it executable and put it on your `PATH`:

```sh
chmod +x bight-darwin-arm64
mv bight-darwin-arm64 /usr/local/bin/bight
```

Each release includes a `checksums.txt` for verification:

```sh
sha256sum -c checksums.txt --ignore-missing
```

**Trying a PR preview:**

For any pushed commit (including from a fork), you can install directly from source:

```sh
go install github.com/AndrewADev/bight@<commit-sha-or-branch>
```

For users without a Go toolchain, a maintainer can run the `Preview` workflow on the PR (Actions → Preview → Run workflow → enter PR number). Binaries for each platform — plus a `checksums.txt` — are then attached to the run as artifacts, downloadable from the run page by anyone with read access to the repo. Preview binaries report a version like `v0.0.0-preview-pr<N>-<sha>` so they can't be mistaken for a release.

## Getting started

### 1. Install

Run once per repo, after cloning:

```sh
bight install
```

This writes the git hook *and* walks you through creating a `.bight.yml` config:

```
bight: hook installed
bight: no config file found. Create .bight.yml? [Y/n]
  Project name [myapp]:
  Env file path [.env]:
  Add env vars to track? [Y/n]
  (blank name to finish)
    Var name: DB_NAME
    Strategy:
      1) template  - interpolate branch/project name (default)
      2) random    - fresh random value on each checkout
    Choice [1]: 1
    Var name:
bight: created .bight.yml
```

### 2. Confirm everything is wired up

```sh
bight doctor
```

```
bight doctor:
  [ok]   git repo detected
  [ok]   config: .bight.yml loaded
  [ok]   config: project = "myapp", 1 env file(s)
  [ok]   hook: installed
  [ok]   env file: .env
  [ok]   vars: all strategies valid
  [ok]   vars: all triggers valid
```

### 3. Preview before it fires automatically

```sh
bight run --dry-run
# bight (dry-run): .env → DB_NAME=myapp_main
```

No files are touched. When you're happy with what you see, you're done — the hook fires on every checkout from here on.

### 4. Switch branches

```sh
git checkout -b feat-login
# bight: .env → DB_NAME=myapp_feat-login
```

## Reference

### Config file

`bight install` generates a starter config, but you can hand-edit `.bight.yml` at any time:

```yaml
project: myapp

defaults:
  branch_template: "{{.Project}}_{{.Branch}}"  # used by the template strategy

env_files:
  - path: .env
    backup: true             # write .env.bak before patching (optional, default false)
    vars:
      - name: DB_NAME
        strategy: template   # renders to e.g. myapp_feat-login
        on: checkout
      - name: JWT_SECRET
        strategy: random     # fresh 64-char hex string on every branch switch
        on: checkout
        sensitive: true      # mask value in console output
```

### Using a non-default config file

If your config isn't named `.bight.yml` or lives at a non-standard path, use `--config`:

```sh
bight run --config path/to/custom.bight.yml
bight doctor --config path/to/custom.bight.yml
```

`--config` is a global flag — it works with any subcommand that reads config.

Alternatively, set the `BIGHT_CONFIG` environment variable. This is convenient for the `post-checkout` hook itself or CI jobs where passing a flag is awkward:

```sh
export BIGHT_CONFIG=path/to/custom.bight.yml
bight doctor   # picks up BIGHT_CONFIG automatically
```

Precedence: `--config` > `BIGHT_CONFIG` > auto-discovery of `.bight.yml` in the current directory. If `BIGHT_CONFIG` points to a missing or unreadable file, `bight` will error rather than silently falling back to auto-discovery. `bight doctor` reports which source was used.

### Manual patching

To apply env patching for the current branch without switching:

```sh
bight run
```

**Tip:** to test how another branch would be patched, suppress the hook when switching so `bight` doesn't fire automatically, then use `--dry-run`:

```sh
git -c core.hooksPath=/dev/null checkout other-branch
bight run --dry-run
```

### Strategies

| Strategy | Output | Typical use |
|---|---|---|
| `template` | Rendered from `{{.Project}}` / `{{.Branch}}` | `DB_NAME` |
| `random` | Fresh 32-byte hex string | `JWT_SECRET`, tokens |
| `deterministic` | Stable 64-char hex derived from project + branch | `DB_NAME` (same value across machines) |

### Sensitive vars (`sensitive`)

Mark a var `sensitive: true` to prevent its value from appearing in console output. The value is still written to the `.env` file normally — only the terminal display is affected.

```yaml
- name: JWT_SECRET
  strategy: random
  on: checkout
  sensitive: true
```

Output with `sensitive: true`:
```
bight: .env → JWT_SECRET=***
```

### Backup files (`backup`)

Set `backup: true` on an env file entry to write a copy of the file to `{path}.bak` before each patch is applied. Useful for inspecting what changed or recovering a previous value.

```yaml
env_files:
  - path: .env
    backup: true
    vars:
      - name: DB_NAME
        strategy: template
        on: checkout
```

The backup is a verbatim copy of the file as it was immediately before patching. It is overwritten on each checkout — only the most recent pre-patch state is kept.

### Preserving comments (`collect-comments`)

Full comment preservation is not supported, as the package we use, `godotenv`, strips comments on rewrite. As a partial workaround, `defaults.collect-comments` re-appends comments collected before the patch was applied:

> **Note:** This is a best-effort feature. Comments are collected from the file before patching and re-appended at the end afterwards — their original positions are not restored, and inline comments (`KEY=val # note`) are lost entirely.

| Value | Behavior |
|---|---|
| `all` | Re-appends every full-line comment |
| `blocks-only` | Re-appends only contiguous comment blocks (≥ 2 lines) — skips isolated `# notes` |
| unset / `none` | Comments are not preserved (default) |

```yaml
defaults:
  collect-comments: blocks-only
```

Comments are always written after the key=value pairs.

### Triggers (`on`)

| Value | When |
|---|---|
| `checkout` | Every branch switch |

### Global config (`~/.bight.yml`)

Settings in `~/.bight.yml` apply across all repos and are overridden field-by-field by the repo's `.bight.yml`. Only `defaults` fields are supported globally — `env_files` and `vars` must be defined in the repo config. If a repo has no `.bight.yml`, `bight` does nothing — the global config alone is not enough to trigger patching.

```yaml
defaults:
  branch_template: "{{.Project}}_{{.Branch}}"
  collect-comments: blocks-only
```

## Developing

See [DEVELOPING.md](./DEVELOPING.md)