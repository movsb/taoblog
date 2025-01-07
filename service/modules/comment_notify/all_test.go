package comment_notify

import (
	"io"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
)

func TestTemplates(t *testing.T) {
	utils.Must(adminTmpl.Execute(io.Discard, AdminData{}))
	utils.Must(guestTmpl.Execute(io.Discard, GuestData{}))
	utils.Must(chanifyTmpl.Execute(io.Discard, AdminData{}))
}
