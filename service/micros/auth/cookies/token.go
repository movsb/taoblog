package cookies

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

const TokenName = `token`

func TokenValue(userID int, password string) string {
	return fmt.Sprintf(`%d:%s`, userID, shasum(password))
}

func AuthorizationValue(userID int, password string) string {
	return TokenName + ` ` + TokenValue(userID, password)
}

// a: token id:sum
// returns: [<id>,<id:sum>]
func ParseAuthorization(a string) (int, string, bool) {
	splits := strings.Fields(a)
	if len(splits) != 2 {
		return 0, "", false
	}
	if splits[0] != TokenName {
		return 0, "", false
	}
	token := splits[1]
	splits = strings.Split(token, `:`)
	if len(splits) != 2 {
		return 0, "", false
	}
	id, err := strconv.Atoi(splits[0])
	if err != nil {
		log.Println(err)
		return 0, "", false
	}

	return id, token, true
}
