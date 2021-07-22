package metrics

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRegistry(t *testing.T) {
	t.SkipNow()
	r := NewRegistry(context.TODO())
	m := http.NewServeMux()
	m.Handle(`/metrics`, r.Handler())
	s := httptest.NewServer(m)
	defer s.Close()
	fmt.Println(s.URL)
	time.Sleep(time.Minute)
}
