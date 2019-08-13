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

	"github.com/danielkvist/whisperer/client"
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
	rDelay     bool
	urls       string
	verbose    bool
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&agent, "agent", "a", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:67.0) Gecko/20100101 Firefox/67.0", "user agent")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "prints error messages")
	rootCmd.PersistentFlags().DurationVarP(&delay, "delay", "d", 1*time.Second, "delay between requests")
	rootCmd.PersistentFlags().IntVarP(&goroutines, "goroutines", "g", 1, "number of goroutines")
	rootCmd.PersistentFlags().StringVarP(&proxy, "proxy", "p", "", "proxy URL")
	rootCmd.PersistentFlags().BoolVarP(&rDelay, "random", "r", false, "random delay between requests")
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

		c, err := client.New(client.WithProxy(proxy), client.WithTimeout(timeout))
		if err != nil {
			log.Fatal(err)
		}

		sema := make(chan struct{}, goroutines)
		seed := rand.NewSource(time.Now().Unix())
		r := rand.New(seed)
		for {
			sema <- struct{}{}
			i := r.Intn(len(sites))
			s := sites[i]

			d := randomDelay(delay, rDelay)
			go visit(s, c, agent, d, verbose, debug, sema, l)
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

func visit(site string, c *http.Client, agent string, delay time.Duration, verbose bool, debug bool, sema <-chan struct{}, l *logger.Logger) {
	time.Sleep(delay)
	defer func() { <-sema }()

	code, _, err := request(c, site, agent)
	if err != nil {
		if debug {
			log.Printf("while making a request: %v", err)
		}

		return
	}

	if verbose {
		l.Println(site + " - " + code)
	}
}

func request(c *http.Client, url string, agent string) (string, int, error) {
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

func randomDelay(delay time.Duration, randomDelay bool) time.Duration {
	if !randomDelay {
		return delay
	}

	r := rand.Intn(int(delay))
	return time.Duration(r)
}
