package cookies

import (
	"testing"
	"time"
)

func TestCookieGen(t *testing.T) {
	ua := `taoblog-ios-client/1.0`
	v := CookieValue(ua, `127.0.0.1`, 2, `b`)
	valid, refresh := ValidateCookieValue(v, ua, `127.0.0.1`, func(userID int) (password string) {
		if userID == 2 {
			return `b`
		}
		return ``
	})
	t.Log(`cookie:`, v)
	if !valid {
		t.Fatal(valid, refresh)
	}
}

func TestCookieExpiration(t *testing.T) {
	const ua = "test-agent"
	const password = "test-password"
	now := time.Now().Truncate(time.Second)
	tests := []struct {
		name    string
		age     time.Duration
		valid   bool
		refresh bool
	}{
		{"new", 0, true, false},
		{"before half lifetime", maxAge/2 - time.Second, true, false},
		{"refresh required", maxAge/2 + time.Second, true, true},
		{"before expiration", maxAge - time.Second, true, true},
		{"at expiration", maxAge, false, false},
		{"expired", maxAge + time.Second, false, false},
		{"future timestamp", -time.Hour, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := cookieValue(ua, "127.0.0.1", 2, password, now.Add(-tt.age))
			valid, refresh := ValidateCookieValue(value, ua, "127.0.0.1", func(id int) string {
				if id != 2 {
					t.Fatalf("unexpected user ID: %d", id)
				}
				return password
			})
			if valid != tt.valid || refresh != tt.refresh {
				t.Fatalf("got (%v, %v), want (%v, %v)", valid, refresh, tt.valid, tt.refresh)
			}
		})
	}
}
