module github.com/movsb/taoblog

require (
	github.com/creack/pty v1.1.9
	github.com/go-sql-driver/mysql v1.4.1
	github.com/golang/protobuf v1.4.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/grpc-gateway v1.14.6
	github.com/movsb/alioss v0.0.0-20180411084708-ae700d1e4460
	github.com/movsb/google-idtoken-verifier v0.0.0-20190329202541-1a6aa2c7e316
	github.com/movsb/taorm v0.0.0-20200705123332-5667be3d9d3c
	github.com/pkg/errors v0.8.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/yuin/goldmark v1.1.30
	go.uber.org/zap v1.10.0
	golang.org/x/crypto v0.0.0-20191206172530-e9b2fee46413
	golang.org/x/sys v0.0.0-20200420163511-1957bb5e6d1f // indirect
	google.golang.org/genproto v0.0.0-20200513103714-09dca8ec2884
	google.golang.org/grpc v1.29.1
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.2.5
)

go 1.13

replace github.com/yuin/goldmark => github.com/movsb/goldmark v1.1.31-0.20200522174842-bc0b03f265ac
