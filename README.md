# bight

A utility to help manage updating .env files when switching branches.

## Usage

### 1. Install the git hook

Run once per repo, after cloning:

```sh
bight install
```

This writes `.git/hooks/post-checkout` pointing at the current `bight` binary.

### 2. Add a config file

Create `.bight.yml` in the repo root:

```yaml
project: myapp

defaults:
  branch_template: "{{.Project}}_{{.Branch}}"  # used by the template strategy

env_files:
  - path: .env
    vars:
      - name: DB_NAME
        strategy: template   # renders to e.g. myapp_feat-login
        on: checkout
      - name: JWT_SECRET
        strategy: random     # fresh 64-char hex string on every branch switch
        on: checkout
```

### 3. Verify your setup

```sh
bight doctor
```

Checks that the git hook is installed, `.bight.yml` (or `.bight.yaml`) is valid, env files exist, and all strategies and triggers are recognized. Run this after cloning or if something isn't patching as expected.

### 4. Switch branches

```sh
git checkout -b feat-login
# bight: .env → DB_NAME=myapp_feat-login
# bight: .env → JWT_SECRET=3f9a...
```

`bight` patches only the listed vars — the rest of your `.env` is left untouched.

### Manual patching and dry runs

To apply env patching for the current branch without switching:

```sh
bight run
```

To preview what would be written without touching any files:

```sh
bight run --dry-run
# bight (dry-run): .env → DB_NAME=myapp_feat-login
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

## Developing

See [DEVELOPING.md](./DEVELOPING.md)