package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	agent      string
	delay      time.Duration
	goroutines int
	timeout    time.Duration
	urls       string
	verbose    bool
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&agent, "agent", "a", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:67.0) Gecko/20100101 Firefox/67.0", "user agent")
	rootCmd.PersistentFlags().DurationVarP(&delay, "delay", "d", 1*time.Second, "delay between requests")
	rootCmd.PersistentFlags().IntVarP(&goroutines, "goroutines", "g", 1, "number of goroutines")
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

		client := &http.Client{Timeout: timeout}
		sema := make(chan struct{}, goroutines)
		seed := rand.NewSource(time.Now().Unix())
		r := rand.New(seed)
		for {
			sema <- struct{}{}
			i := r.Intn(len(sites) - 1)
			s := sites[i]

			go func(site string) {
				defer func() {
					time.Sleep(delay)
					<-sema
				}()

				status, err := request(client, site)
				if err != nil {
					log.Printf("while making a request for %v: %v", site, err)
					return
				}

				if verbose {
					log.Printf("visited %v - %v", site, status)
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

func request(c *http.Client, url string) (int, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", agent)

	resp, err := c.Do(req)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}
