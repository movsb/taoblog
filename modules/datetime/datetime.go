package datetime

import (
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
)

const (
	mysqlFormat   = "2006-01-02 15:04:05"
	httpgmtFormat = "Mon, 02 Jan 2006 15:04:05 GMT"
	feedFormat    = "Mon, 02 Jan 2006 15:04:05"
)

// MyLocal is the MySQL format of Local time.
func MyLocal() string {
	return time.Now().Format(mysqlFormat)
}

func ProtoLocal() *timestamp.Timestamp {
	t := time.Now()
	return &timestamp.Timestamp{
		Seconds: t.Unix(),
		Nanos:   int32(t.Nanosecond()),
	}
}

// My2Gmt returns the GMT representation of MySQL local time.
func My2Gmt(t string) string {
	tm, _ := time.ParseInLocation(mysqlFormat, t, time.Local)
	return tm.UTC().Format(httpgmtFormat)
}

// My2Feed returns the feed time of MySQL local time.
func My2Feed(date string) string {
	tm, _ := time.ParseInLocation(mysqlFormat, date, time.Local)
	return tm.Format(feedFormat)
}

// My2Proto ...
func My2Proto(t string) *timestamp.Timestamp {
	tm, _ := time.ParseInLocation(mysqlFormat, t, time.Local)
	return &timestamp.Timestamp{
		Seconds: tm.Unix(),
		Nanos:   int32(tm.Nanosecond()),
	}
}

// Proto2My ...
func Proto2My(t timestamp.Timestamp) string {
	tm := time.Unix(t.Seconds, int64(t.Nanos))
	return tm.Format(mysqlFormat)
}
