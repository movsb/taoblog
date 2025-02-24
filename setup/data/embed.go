package setup_data

import (
	"embed"
)

//go:embed posts.sql files.sql
var Root embed.FS
