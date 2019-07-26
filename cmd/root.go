package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/danielkvist/whisperer/logger"

	"github.com/spf13/cobra"
)

var (
	agent      string
	debug      bool
	delay      time.Duration
	goroutines int
	timeout    time.Duration
	proxy      string
	urls       string
	verbose    bool
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&agent, "agent", "a", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:67.0) Gecko/20100101 Firefox/67.0", "user agent")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "prints error messages")
	rootCmd.PersistentFlags().DurationVarP(&delay, "delay", "d", 1*time.Second, "delay between requests")
	rootCmd.PersistentFlags().IntVarP(&goroutines, "goroutines", "g", 1, "number of goroutines")
	rootCmd.PersistentFlags().StringVarP(&proxy, "proxy", "p", "", "proxy URL")
	rootCmd.PersistentFlags().DurationVarP(&timeout, "timeout", "t", 3*time.Second, "max time to wait for a response before canceling the request")
	rootCmd.PersistentFlags().StringVar(&urls, "urls", "./urls.txt", "simple .txt file with URL's to visit")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enables verbose mode")
}

var rootCmd = &cobra.Command{
	Use:   "whisperer",
	Short: "whisperer makes HTTP request constantly in order to generate random HTTP/DNS traffic noise.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Open(urls)
		if err != nil {
			return err
		}
		defer f.Close()

		sites, err := readURLS(f)
		if err != nil {
			return fmt.Errorf("while reading URLs from %q: %v", urls, err)
		}

		if len(sites) == 0 {
			return fmt.Errorf("there is no valid URL in the file %v", urls)
		}

		l := logger.New(os.Stdout, goroutines)
		if !verbose {
			l.Stop()
		}

		client := &http.Client{Timeout: timeout}
		if proxy != "" {
			if err := clientWithProxy(client, proxy); err != nil {
				log.Fatal(err)
			}
		}

		sema := make(chan struct{}, goroutines)
		seed := rand.NewSource(time.Now().Unix())
		r := rand.New(seed)
		for {
			sema <- struct{}{}
			i := r.Intn(len(sites))
			s := sites[i]

			go func(site string) {
				defer delayRequest(delay, sema)
				code, _, err := request(client, site)
				if err != nil {
					if debug {
						log.Printf("while making a request for %v: %v", site, err)
					}
					return
				}

				if verbose {
					l.Println(site + " - " + code)
				}
			}(s)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func readURLS(r io.Reader) ([]string, error) {
	urls := []string{}
	input := bufio.NewScanner(r)
	for input.Scan() {
		url := input.Text()
		if url == "" {
			continue
		}

		if !strings.HasPrefix(url, "https://") {
			url = "https://" + url
		}
		urls = append(urls, url)
	}

	return urls, input.Err()
}

func clientWithProxy(c *http.Client, proxy string) error {
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return err
	}

	tr := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	c.Transport = tr
	return nil
}

func delayRequest(d time.Duration, sema <-chan struct{}) {
	time.Sleep(d)
	<-sema
}

func request(c *http.Client, url string) (string, int, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("User-Agent", agent)

	resp, err := c.Do(req)
	if err != nil {
		return "", 0, err
	}

	return resp.Status, resp.StatusCode, nil
}
