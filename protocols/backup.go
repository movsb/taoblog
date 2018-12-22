package protocols

import "io"

type GetBackupRequest struct {
	W io.Writer
}
