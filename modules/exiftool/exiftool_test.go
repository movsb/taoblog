package exiftool

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestCommandContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := CommandContext(ctx, nil, "-ver"); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
}

func TestCommandEmptyImage(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if _, err := Command(strings.NewReader(""), "-json", "-"); err == nil {
		t.Fatal("empty image should return an error")
	}
}
