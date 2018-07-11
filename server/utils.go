package main

import (
	"fmt"
	"net/http"
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

// EndReq succs or fails a request.
func EndReq(c *gin.Context, err interface{}, dat interface{}) {
	succ := false

	if err == nil {
		succ = true
	} else if val, ok := err.(error); ok {
		succ = val == nil
	} else if val, ok := err.(bool); ok {
		succ = val
	} else if val, ok := err.(string); ok {
		succ = val == ""
	}

	if succ {
		c.JSON(http.StatusOK, dat)
	} else {
		c.JSON(http.StatusInternalServerError, err)
	}
}

// BuildQueryString builds SQL query string.
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
