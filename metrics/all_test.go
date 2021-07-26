package metrics

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mssola/user_agent"
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

func TestUserAgent(t *testing.T) {
	ua := user_agent.New(`Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:90.0) Gecko/20100101 Firefox/90.0`)
	fmt.Println(ua.Bot())
	fmt.Println(ua.Browser())
	fmt.Println(ua.Engine())
	fmt.Println(ua.Localization())
	fmt.Println(ua.Mobile())
	fmt.Println(ua.Mozilla())
	fmt.Println(ua.OS())
	fmt.Println(ua.OSInfo())
	fmt.Println(ua.Platform())
	fmt.Println(ua.UA())
}
