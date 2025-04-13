package setup_data

import (
	"embed"
)

//go:embed *.sql
var Root embed.FS
