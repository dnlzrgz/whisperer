package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
			"with http",
			"http://amazon.com",
			[]string{"http://amazon.com"},
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

func TestRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		switch r.URL.String() {
		case "/ok":
			w.WriteHeader(http.StatusOK)
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
		case "/404":
			w.WriteHeader(http.StatusNotFound)
		}
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	tt := []struct {
		url            string
		expectedStatus string
	}{
		{
			ts.URL + "/ok",
			"200 OK",
		},
		{
			ts.URL + "/error",
			"500 Internal Server Error",
		},
		{
			ts.URL + "/404",
			"404 Not Found",
		},
	}

	client := &http.Client{}
	for _, tc := range tt {
		t.Run(tc.url, func(t *testing.T) {
			status, err := request(client, tc.url, "")
			if err != nil {
				t.Fatalf("while making a request to %v: %v", tc.url, err)
			}

			if status != tc.expectedStatus {
				t.Fatalf("expected status code to be %v. got=%v", tc.expectedStatus, status)
			}
		})
	}

}
