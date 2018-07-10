package main

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

func joinInts(ints []int64, delim string) string {
	return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ints)), delim), "[]")
}

func strInSlice(str []string, s string) bool {
	for _, b := range str {
		if b == s {
			return true
		}
	}
	return false
}

func finishDone(c *gin.Context, code int, msgs string, data interface{}) {
	c.JSON(200, &xJSONRet{code, msgs, data})
}

func finishError(c *gin.Context, code int, err error) {
	c.JSON(200, &xJSONRet{code, fmt.Sprint(err), nil})
}

func BuildQueryString(fields map[string]interface{}) string {
	var q string

	q += fmt.Sprintf("SELECT %s FROM %s", fields["select"], fields["from"])

	if where, ok := fields["where"]; ok {
		switch typed := where.(type) {
		case []string:
			clause := ""
			for _, w := range typed {
				clause += " AND (" + w + ")"
			}
			if clause != "" {
				q += " WHERE 1" + clause
			}
		default:
			panic("invalid where")
		}
	}

	if groupby, ok := fields["groupby"]; ok {
		q += " GROUP BY " + fmt.Sprint(groupby)
	}

	if having, ok := fields["having"]; ok {
		q += " HAVING " + fmt.Sprint(having)
	}

	if orderby, ok := fields["orderby"]; ok {
		q += " ORDER BY " + fmt.Sprint(orderby)
	}

	if limit, ok := fields["limit"]; ok {
		if offset, ok := fields["offset"]; ok {
			q += fmt.Sprintf(" LIMIT %v,%v", limit, offset)
		} else {
			q += fmt.Sprintf(" LIMIT %v", limit)
		}
	}

	return q
}
