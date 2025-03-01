package backups_crypto

import (
	"io"

	"filippo.io/age"
	"github.com/movsb/taoblog/modules/utils"
)

type EncodeDecoder interface {
	Write(p []byte) (int, error)
	Close() error
}

type Age struct {
	io.WriteCloser
}

func NewAge(identity string, w io.Writer) (_ EncodeDecoder, outErr error) {
	defer utils.CatchAsError(&outErr)
	ident := utils.Must1(age.ParseX25519Identity(identity))
	// NOTE 这里使用的是自己，为了安全，可以用不同的钥匙。
	ew := utils.Must1(age.Encrypt(w, ident.Recipient()))
	return &Age{WriteCloser: ew}, nil
}
