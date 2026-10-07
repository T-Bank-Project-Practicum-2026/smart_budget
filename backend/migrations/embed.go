package migrations

import "embed"

// Files contains the SQL migrations shipped with the migration binary.
//
//go:embed *.sql notifications/*.sql
var Files embed.FS
