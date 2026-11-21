package migrations

import "embed"

//go:embed *.sql notifications/*.sql
var Files embed.FS
