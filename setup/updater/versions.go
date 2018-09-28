package main

import (
	"database/sql"
)

// VersionUpdater is
type VersionUpdater struct {
	version string
	updater func(tx *sql.Tx)
}

var gVersions = []VersionUpdater{
	{"1.1.11", v1_1_11},
}
