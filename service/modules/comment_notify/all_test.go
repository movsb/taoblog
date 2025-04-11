package comment_notify

import (
	"io"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
)

func TestTemplates(t *testing.T) {
	utils.Must(authorTmpl.Execute(io.Discard, Data{}))
	utils.Must(guestTmpl.Execute(io.Discard, Data{}))
	utils.Must(chanifyTmpl.Execute(io.Discard, AdminData{}))
}
