package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
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

func CreateHandler(secret string, reloaderPath string, sendNotify func(content string)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		sum := strings.TrimPrefix(r.Header.Get(`X-Hub-Signature-256`), `sha256=`)
		payload, err := decode(r.Body, secret, sum)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		if w := payload.WorkflowRun; w.Status == `completed` {
			switch w.Conclusion {
			case `success`:
				client := http.Client{
					Transport: &http.Transport{
						Dial: func(network, addr string) (net.Conn, error) {
							return net.Dial(`unix`, reloaderPath)
						},
					},
				}
				go func() {
					rsp, err := client.Post(`http://localhost/reload`, "", nil)
					if err != nil {
						log.Println(err)
						return
					}
					if rsp.StatusCode == http.StatusOK {
						log.Println("已成功触发重启。")
						return
					}
					log.Println("未处理的重启失败。", rsp.Status)
				}()
				return
			default:
				sendNotify(fmt.Sprintf("结果未知：%s", w.Conclusion))
			}
		}
	}
}
