# migrations/

This project currently manages the schema with GORM's `AutoMigrate`
(see `internal/database/database.go`), which runs automatically on every
server start and creates/updates tables to match the model structs.

This folder is reserved for versioned SQL migration files if/when the
project moves to an explicit migration tool (e.g. `golang-migrate`,
`goose`) for production — `AutoMigrate` is convenient for development but
doesn't support column renames, data backfills, or safe rollbacks, which
matters once there's real production data.
