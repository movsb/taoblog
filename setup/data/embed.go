package setup_data

import (
	"embed"
)

//go:embed schemas.sqlite.sql
var Root embed.FS
