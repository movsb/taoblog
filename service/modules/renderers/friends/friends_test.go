package friends

import "testing"

func TestResolveURL(t *testing.T) {
	testCases := []struct {
		SiteURL    string
		FaviconURL string
		Want       string
	}{
		{
			SiteURL:    `https://blog.twofei.com`,
			FaviconURL: ``,
			Want:       `https://blog.twofei.com/favicon.ico`,
		},
		{
			SiteURL:    `https://blog.twofei.com`,
			FaviconURL: `https://blog.twofei.com/favicon.avif`,
			Want:       `https://blog.twofei.com/favicon.avif`,
		},
		{
			SiteURL:    `https://blog.twofei.com`,
			FaviconURL: `/favicon.avif`,
			Want:       `https://blog.twofei.com/favicon.avif`,
		},
		{
			SiteURL:    `https://blog.twofei.com`,
			FaviconURL: `favicon.avif`,
			Want:       `https://blog.twofei.com/favicon.avif`,
		},
		{
			SiteURL:    `https://blog.twofei.com/sub`,
			FaviconURL: `favicon.avif`,
			Want:       `https://blog.twofei.com/favicon.avif`,
		},
		{
			SiteURL:    `https://blog.twofei.com/sub`,
			FaviconURL: `/favicon.avif`,
			Want:       `https://blog.twofei.com/favicon.avif`,
		},
		{
			SiteURL:    `https://blog.twofei.com/sub/`,
			FaviconURL: `favicon.avif`,
			Want:       `https://blog.twofei.com/sub/favicon.avif`,
		},
		{
			SiteURL:    `https://blog.twofei.com/sub/`,
			FaviconURL: `/favicon.avif`,
			Want:       `https://blog.twofei.com/favicon.avif`,
		},
	}

	for i, tc := range testCases {
		u, err := resolveIconURL(tc.SiteURL, tc.FaviconURL)
		if err != nil || u != tc.Want {
			t.Errorf(`error: %d: %s %s %s`, i, tc.SiteURL, tc.FaviconURL, tc.Want)
			continue
		}
	}
}
