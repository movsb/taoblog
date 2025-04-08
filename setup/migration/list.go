package migration

import (
	"database/sql"
)

// VersionUpdater is
type VersionUpdater struct {
	version int
	update  func(tx *sql.Tx)
}

var gVersions = []VersionUpdater{
	{0, v0},
	{1, v1},
	{2, v2},
	{3, v3},
	{4, v4},
	{5, v5},
	{6, v6},
	{7, v7},
	{8, v8},
	{9, v9},
	{10, v10},
	{11, v11},
	{12, v12},
	{13, v13},
	{14, v14},
	{15, v15},
	{16, v16},
	{17, v17},
	{18, v18},
	{19, v19},
	{20, v20},
	{21, v21},
	{22, v22},
	{23, v23},
	{24, v24},
	{25, v25},
	{26, v26},
	{27, v27},
	{28, v28},
	{29, v29},
	{30, v30},
	{31, v31},
	{32, v32},
	{33, v33},
	{34, v34},
	{35, v35},
	{36, v36},
	{37, v37},
	{38, v38},
	{39, v39},
	{40, v40},
	{41, v41},
	{42, v42},
}

// MaxVersionNumber ...
func MaxVersionNumber() int {
	return gVersions[len(gVersions)-1].version
}
