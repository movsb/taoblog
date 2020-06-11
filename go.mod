module github.com/movsb/taoblog

require (
	github.com/creack/pty v1.1.9
	github.com/gin-gonic/gin v1.4.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/golang/protobuf v1.4.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/grpc-gateway v1.14.6
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/movsb/alioss v0.0.0-20180411084708-ae700d1e4460
	github.com/movsb/google-idtoken-verifier v0.0.0-20190329202541-1a6aa2c7e316
	github.com/movsb/taorm v0.0.0-20200410180644-b357f5988367
	github.com/prometheus/client_golang v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/yuin/goldmark v1.1.30
	golang.org/x/crypto v0.0.0-20191206172530-e9b2fee46413
	google.golang.org/genproto v0.0.0-20200513103714-09dca8ec2884
	google.golang.org/grpc v1.29.1
	gopkg.in/yaml.v2 v2.2.5
)

go 1.13

replace github.com/yuin/goldmark => github.com/movsb/goldmark v1.1.31-0.20200522174842-bc0b03f265ac
