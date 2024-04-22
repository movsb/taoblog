package webhooks

import "testing"

func TestGitHub(t *testing.T) {
	// https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries#testing-the-webhook-payload-validation
	payload := "Hello, World!"
	secret := "It's a Secret to Everybody"
	expected := "757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17"
	if !verifyIntegrity([]byte(payload), secret, expected) {
		t.Fatal(`github validation not equal`)
	}
}
