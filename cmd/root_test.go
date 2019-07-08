package cmd

import (
	"strings"
	"testing"
)

func TestReadURLS(t *testing.T) {
	tt := []struct {
		name   string
		input  string
		output []string
	}{
		{
			"without https",
			"google.com\nbing.com\ndkvist.com",
			[]string{"https://google.com", "https://bing.com", "https://dkvist.com"},
		},
		{
			"with https",
			"https://amazon.com",
			[]string{"https://amazon.com"},
		},
		{
			"empty",
			"",
			[]string{},
		},
		{
			"only a newline",
			"\n",
			[]string{},
		},
		{
			"multiple newlines",
			"\n\n\n\n\n",
			[]string{},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			urls, err := readURLS(strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("while reading URLs: %v", err)
			}

			if len(urls) != len(tc.output) {
				t.Fatalf("expected URLs to have a len of %v. got=%v", len(tc.output), len(urls))
			}

			for i, u := range urls {
				if tc.output[i] != u {
					t.Fatalf("expected URLs to have URL %q. got=%q", tc.output[i], u)
				}
			}
		})
	}
}
