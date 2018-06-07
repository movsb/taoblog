package datetime

import (
	"time"
)

const (
	mysqlFormat   = "2006-01-02 15:04:05"
	httpgmtFormat = "Mon, 02 Jan 2006 15:04:05 GMT"
	feedFormat    = "Mon, 02 Jan 2006 15:04:05"
)

// MyGMT is the UTC time
func MyGmt() string {
	return time.Now().UTC().Format(mysqlFormat)
}

// MyLocal is the Local time
func MyLocal() string {
	return time.Now().Format(mysqlFormat)
}

func My2Gmt(t string) string {
	tim, err := time.Parse(mysqlFormat, t)
	if err != nil {
		panic(err)
	}
	return tim.Format(httpgmtFormat)
}

func Gmt2My(t string) string {
	tim, err := time.Parse(httpgmtFormat, t)
	if err != nil {
		panic(err)
	}
	return tim.Format(mysqlFormat)
}

func My2Local(t string) string {
	tim, err := time.Parse(mysqlFormat, t)
	if err != nil {
		panic(err)
	}
	return tim.Local().Format(mysqlFormat)
}

func Local2My(t string) string {
	tim, err := time.ParseInLocation(mysqlFormat, t, time.Local)
	if err != nil {
		panic(err)
	}
	return tim.UTC().Format(mysqlFormat)
}

func Local2Gmt(t string) string {
	tim, err := time.ParseInLocation(mysqlFormat, t, time.Local)
	if err != nil {
		panic(err)
	}
	return tim.UTC().Format(httpgmtFormat)
}

func Local2Timestamp(t string) int64 {
	tim, err := time.ParseInLocation(mysqlFormat, t, time.Local)
	if err != nil {
		panic(err)
	}
	return tim.Unix()
}

func GmtNow() string {
	return time.Now().UTC().Format(httpgmtFormat)
}

func IsValidMy(t string) bool {
	_, err := time.Parse(mysqlFormat, t)
	return err == nil
}

func FeedNow() string {
	return time.Now().Format(feedFormat)
}

// MonthStartEnd returns
// month: [1,12]
func MonthStartEnd(year int, month int) (string, string) {
	days := [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if year%4 == 0 && year%100 != 0 || year%400 == 0 {
		days[1] = 29
	}

	t1 := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	t2 := time.Date(year, time.Month(month), days[month-1], 23, 59, 59, 999, time.Local)
	return t1.UTC().Format(mysqlFormat), t2.UTC().Format(mysqlFormat)
}

func YearStartEnd(year int) (string, string) {
	t1 := time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local)
	t2 := time.Date(year, time.December, 31, 23, 59, 59, 999, time.Local)
	return t1.UTC().Format(mysqlFormat), t2.UTC().Format(mysqlFormat)
}
