# Changelog
## v0.5.0 - 2026-09-27

### Bug Fixes

- Validate templates (#41)

### Features

- Sanitize function for templates (#40)

## v0.4.0 - 2026-09-26

### Bug Fixes

- Prevent value modification due to escaping (#38)

### Features

- Preserve all full-line comments (#37)

### Miscellaneous

- Include PR numbers in CHANGELOG (#39)

## v0.3.2 - 2026-09-19

### Bug Fixes

- Handle missing binary in hook (#34)
- Switch to maintained yaml package (#36)

### Documentation

- Add GH Pages site w/ llms.txt (#32)

### Miscellaneous

- Upgrade to go 1.26 (#33)

### Refactor

- Apply go fix modernizers (#33)

## v0.3.1 - 2026-06-09

### Bug Fixes

- Skip interactive config init when non-interactive (#29)

### Documentation

- Add usage for all commands (#27)

### Miscellaneous

- Streamline preview builds
- Prevent divergences from changelog (#31)

### Refactor

- Shared confirm prompt (#28)

## v0.3.0 - 2026-06-06

### Features

- Add copy functionality (e.g. seeding/bootstrapping `.env` files) (#23)

## v0.2.0 - 2026-05-28

### Bug Fixes

- Skip files with no updates (#24)

### Features

- Specify config path via env var (#25)

### Miscellaneous

- Add on-demand preview builds (#22)

## v0.1.0 - 2026-04-30

### Bug Fixes

- Grant needed permissions

### Documentation

- Add homebrew formula (#19)

### Features

- Support enabling backup files (#20)

### Miscellaneous

- Ignore local claude files (#18)
- Add Taskfile (#21)

## v0.0.2 - 2026-04-26

### Bug Fixes

- Account for worktree context (#17)
- Prevent partial writes and permissions loss (#16)

### Miscellaneous

- Invoke tap update on new tag (#15)

## v0.0.1 - 2026-04-17

### Bug Fixes

- Graceful no-op when no config
- Branch resolution in worktrees (#14)

### Documentation

- Deterministic strategy (#4)
- Style note and cleanup (#11)
- Update summary and install (#13)

### Features

- Env patching on branch checkout
- Support yaml ending
- Guided config creation
- Add uninstall command
- Add --version flag (#2)
- Add diagnostics (#3)
- Add deterministic strategy
- Add --dry-run flag (#5)
- Partial comment preservation (#6)
- Support global config (#8)
- Support specifying config path (#9)
- Support colorized output (#10)
- Add sensitive flag (#12)

### Miscellaneous

- Init project
- Basic conventions and checks
- Add tagged versions to release (#7)


