module github.com/movsb/taoblog

require (
	github.com/PuerkitoBio/goquery v1.5.1
	github.com/golang/protobuf v1.5.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/litao91/goldmark-mathjax v0.0.0-20191101121019-011def32b12f
	github.com/mattn/go-sqlite3 v1.14.7
	github.com/movsb/google-idtoken-verifier v0.0.0-20190329202541-1a6aa2c7e316
	github.com/movsb/taorm v0.0.0-20201209183410-91bafb0b22a6
	github.com/mssola/user_agent v0.5.3
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/cobra v1.0.0
	github.com/yuin/goldmark v1.2.1
	go.uber.org/zap v1.10.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	google.golang.org/genproto v0.0.0-20210325224202-eed09b1b5210
	google.golang.org/grpc v1.36.1
	google.golang.org/protobuf v1.26.0
	gopkg.in/yaml.v2 v2.3.0
)

go 1.16

replace github.com/yuin/goldmark => github.com/movsb/goldmark v1.4.1-0.20210827190033-0f221783bf61

// replace github.com/movsb/taorm => /Users/tao/code/taorm
