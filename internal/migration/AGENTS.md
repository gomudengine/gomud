# Migration Package Context

## Overview

The `internal/migration` package handles automatic data file migrations when the server version advances past a version boundary. It runs at startup, after config files are loaded, and upgrades on-disk data files to match the current server version.

## Key Components

### Entry Point (`migration.go`)

- **`Run(lastConfigVersion, serverVersion version.Version) error`**: Main entry point called at server startup.
  - If `lastConfigVersion == serverVersion`, returns immediately (no-op).
  - Creates a full backup of the datafiles directory before applying any migrations.
  - Calls `doAllMigrations` with the last known version.
  - On success, updates `Server.CurrentVersion` in config via `configs.SetVal`.
  - On failure, restores the backup before returning the error.
- **`doAllMigrations(lastConfigVersion version.Version) error`**: Runs each migration function in chronological order, gated by `version.IsOlderThan`.

### Backup System (`backup.go`)

- **`datafilesBackup() (string, error)`**: Creates a temporary directory copy of the entire datafiles folder using `os.MkdirTemp`. Returns the temp dir path. The caller is responsible for cleanup (`defer os.RemoveAll`).
- **`copyDir(src, dst string) error`**: Recursive directory copy using `filepath.WalkDir`.
- **`copyFile(srcFile, dstFile string) error`**: Single-file copy via `io.Copy`.

### Version-Specific Migrations

#### `0.9.1.go` — Room Zone Config Migration

- **`migrate_RoomZoneConfig() error`**: Migrates `ZoneConfig` data out of individual room YAML files into per-zone `zone-config.yaml` files.
  - Scans all `DATAFILES/rooms/*/*.yaml` files matching the `###.yaml` pattern.
  - Skips rooms whose zone directory already has a `zone-config.yaml`.
  - For rooms containing a `zoneconfig:` YAML key, extracts the data, writes it to `zone-config.yaml`, then rewrites the room file without the `zoneconfig` key.
  - Uses a local copy of the `zoneConfig_1_0_0` struct (snapshot of the struct as of that version) to avoid being affected by future struct changes.
  - Logs progress via `mudlog.Info`.

## Adding New Migrations

1. Create a new file named after the target version (e.g., `1.2.0.go`).
2. Implement a `migrate_<Description>() error` function.
3. Add a call to it in `doAllMigrations` gated with `lastConfigVersion.IsOlderThan(version.New(...))`, in chronological order after existing migrations.

## Dependencies

- `internal/configs`: Reading datafile paths and writing the updated version
- `internal/version`: Version comparison for migration gating
- `internal/rooms`: Unmarshalling room structs for canonical re-serialization
- `internal/mudlog`: Logging migration progress
- `gopkg.in/yaml.v2`: YAML parsing and serialization of room/zone files
- Standard library: `os`, `io`, `path/filepath`, `regexp`, `strings`

## Special Considerations

- **Backup-then-restore pattern**: All migrations are wrapped in a backup. If any migration fails, the datafiles directory is restored to the pre-migration state.
- **Struct snapshots**: Migration code uses local copies of structs as they existed at the migration target version, not the current live structs, to ensure correctness when migrating old data.
- **Idempotency**: Each migration checks whether the target state already exists (e.g., `zone-config.yaml` present) before acting, making migrations safe to re-run.
- **Config not backed up**: The backup covers the datafiles directory only. Migrations that modify config files need special handling.
