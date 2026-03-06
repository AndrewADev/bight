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

### 3. Switch branches

```sh
git checkout -b feat-login
# bight: .env → DB_NAME=myapp_feat-login
# bight: .env → JWT_SECRET=3f9a...
```

`bight` patches only the listed vars — the rest of your `.env` is left untouched.

### Strategies

| Strategy | Output | Typical use |
|---|---|---|
| `template` | Rendered from `{{.Project}}` / `{{.Branch}}` | `DB_NAME` |
| `random` | Fresh 32-byte hex string | `JWT_SECRET`, tokens |

### Triggers (`on`)

| Value | When |
|---|---|
| `checkout` | Every branch switch |

## Developing

See [DEVELOPING.md](./DEVELOPING.md)