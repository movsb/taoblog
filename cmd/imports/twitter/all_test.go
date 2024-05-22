package twitter

import (
	"log"
	"os"
	"testing"
)

func TestParseTweets(t *testing.T) {
	t.SkipNow()
	tweets, err := ParseTweets(os.DirFS(`/Users/tao/Downloads/twitter-2024-05-20-bd63b478db520d4d86858aa7fcf6f326834edc48f15ec461cd2c135e2835616d`))
	if err != nil {
		t.Fatal(err)
	}

	// enc := json.NewEncoder(os.Stdout)
	// enc.SetEscapeHTML(false)
	// enc.SetIndent(``, `  `)
	// enc.Encode(tweets)

	n := 0
	m := 0
	for _, t := range tweets {
		if t.IsSelfTweet() {
			n++
		} else {
			m++
		}
	}
	t.Log(n, m)

	for _, t := range tweets {
		log.Println(t.Assets(false))
	}
}
