package datetime

import (
	"time"
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
