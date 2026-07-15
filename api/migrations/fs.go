package migrations

import "embed"

// FS keeps schema changes inside the API binary so every process uses the same migrations.
//
//go:embed *.sql
var FS embed.FS
