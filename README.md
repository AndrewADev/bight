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

Checks that the git hook is installed, `.bight.yml` (or `.bight.yaml`) is valid, env files exist, and all strategies and triggers are recognised. Run this after cloning or if something isn't patching as expected.

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

### Triggers (`on`)

| Value | When |
|---|---|
| `checkout` | Every branch switch |

## Developing

See [DEVELOPING.md](./DEVELOPING.md)