package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
)

func verifyIntegrity(content []byte, secret string, expected string) bool {
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write(content)
	sum := hash.Sum(nil)
	return fmt.Sprintf(`%x`, sum) == expected
}

const maxBodySize = 25 << 20

type Payload struct {
	WorkflowRun struct {
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
	} `json:"workflow_run"`
}

func decode(r io.Reader, secret string, sum string) (*Payload, error) {
	all, err := io.ReadAll(io.LimitReader(r, maxBodySize))
	if err != nil {
		return nil, err
	}

	if !verifyIntegrity(all, secret, sum) {
		return nil, fmt.Errorf(`github checksum error`)
	}

	p := Payload{}
	if err := json.Unmarshal(all, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
