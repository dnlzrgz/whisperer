// Package client provides utilities to create easily new http.Client.
package client

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Option defines an option for a new
// *http.Client.
type Option func(c *http.Client) error

// WithTimeout receives a timeout and returns
// an Option that applies it to a new *http.Client.
func WithTimeout(t time.Duration) Option {
	return func(c *http.Client) error {
		c.Timeout = t
		return nil
	}
}

// WithProxy receives a proxy's URL and returns an
// Option function that parses that URL and, if there is
// no errors, creates a new http.Transport configured
// that is applied to a new *http.Client.
func WithProxy(proxy string) Option {
	return func(c *http.Client) error {
		if proxy == "" {
			c.Transport = nil
			return nil
		}

		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return err
		}

		tr := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		c.Transport = tr
		return nil
	}
}

// New returns a new *http.Client or an error after applying
// the received Options.
func New(opts ...Option) (*http.Client, error) {
	c := &http.Client{}

	for _, option := range opts {
		err := option(c)
		if err != nil {
			return nil, fmt.Errorf("while creating a new http.Client: %v", err)
		}
	}

	return c, nil
}
