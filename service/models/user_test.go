package models

import (
	"bytes"
	"encoding/json"
	"testing"
)

// v0.10.2 stored the attestation format in attestationType.
func TestCredentialsLegacyJSON(t *testing.T) {
	const legacy = `[{"id":"AQID","publicKey":"BAUG","attestationType":"none","transport":["internal"],"flags":{"userPresent":true,"userVerified":true,"backupEligible":true,"backupState":true},"authenticator":{"AAGUID":"AAAAAAAAAAAAAAAAAAAAAA==","signCount":7,"cloneWarning":false,"attachment":"platform"}}]`
	for _, input := range []any{legacy, []byte(legacy)} {
		var credentials Credentials
		if err := credentials.Scan(input); err != nil {
			t.Fatal(err)
		}
		if len(credentials) != 1 {
			t.Fatalf("got %d credentials", len(credentials))
		}
		c := credentials[0]
		if !bytes.Equal(c.ID, []byte{1, 2, 3}) || !bytes.Equal(c.PublicKey, []byte{4, 5, 6}) || c.Authenticator.SignCount != 7 || !c.Flags.BackupEligible || !c.Flags.BackupState {
			t.Fatalf("legacy credential data changed: %+v", c)
		}
		if c.AttestationFormat != "none" || c.AttestationType != "" {
			t.Fatalf("legacy attestation format was not migrated: %+v", c)
		}
		stored, err := credentials.Value()
		if err != nil {
			t.Fatal(err)
		}
		var restored Credentials
		if err := restored.Scan(stored); err != nil {
			t.Fatal(err)
		}
		before, _ := json.Marshal(credentials)
		after, _ := json.Marshal(restored)
		if !bytes.Equal(before, after) {
			t.Fatalf("credential changed after saving: %s != %s", before, after)
		}
	}
}
