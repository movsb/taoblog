package whois

import (
	"testing"
)

func TestQueryDomain(t *testing.T) {
	kv, err := QueryDomainExpiration(`example.com`)
	t.Log(kv, err)
}
