package main

import (
	"database/sql"
)

// VersionUpdater is
type VersionUpdater struct {
	version int
	updater func(tx *sql.Tx)
}

var gVersions = []VersionUpdater{
	{0, v0},
	{1, v1},
}
