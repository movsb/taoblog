package datetime

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	fmt.Println(MyGmt())
	fmt.Println(MyLocal())
	fmt.Println(My2Gmt("2018-06-07 13:00:27"))
	fmt.Println(Gmt2My("Thu, 07 Jun 2018 13:00:27 GMT"))
	fmt.Println(My2Local("2018-06-07 13:00:27"))
	fmt.Println(Local2My("2018-06-07 13:00:27"))
	fmt.Println(Local2Gmt("2018-06-07 13:00:27"))
	fmt.Println(Local2Timestamp("2018-06-07 13:00:27"))
	fmt.Println(GmtNow())
	fmt.Println(IsValidMy("2018-06-07 13:00:27"))
	fmt.Println(FeedNow())
	fmt.Println(YearStartEnd(2018))
	fmt.Println(MonthStartEnd(2017, 3))
}
