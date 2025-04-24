package gravatar

import (
	"context"
	"testing"
)

func TestGet(t *testing.T) {
	t.SkipNow()
	get(context.Background(), `https://cdn.sep.cc/avatar/`, `92e6ba14432b031c166ac8646b95fc8dc68deb42bc3791e86cd2d01ac6fcb7e1`)
}
