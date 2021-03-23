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

	"github.com/spf13/cobra"
)

func Root() *cobra.Command {
	var debug bool
	var agent string
	var delay time.Duration
	var goroutines int
	var timeout time.Duration
	var proxy string
	var rDelay bool
	var urls string
	var verbose bool

	root := &cobra.Command{
		Use:   "whisperer",
		Short: "whisperer makes HTTP request constantly in order to generate random HTTP/DNS traffic noise.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if goroutines <= 0 {
				return fmt.Errorf("number of goroutines cannot be less or equal to 0")
			}

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
				return fmt.Errorf("there is no valid URLs in the file %v", urls)
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
				go visit(s, c, agent, d, verbose, debug, sema)
			}
		},
		SilenceUsage: true,
	}

	root.Flags().StringVarP(&agent, "agent", "a", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:67.0) Gecko/20100101 Firefox/67.0", "user agent")
	root.Flags().BoolVar(&debug, "debug", false, "prints error messages")
	root.Flags().DurationVarP(&delay, "delay", "d", 1*time.Second, "delay between requests")
	root.Flags().IntVarP(&goroutines, "goroutines", "g", 1, "number of goroutines")
	root.Flags().StringVarP(&proxy, "proxy", "p", "", "proxy URL")
	root.Flags().BoolVarP(&rDelay, "random", "r", false, "random delay between requests")
	root.Flags().DurationVarP(&timeout, "timeout", "t", 3*time.Second, "max time to wait for a response before canceling the request")
	root.Flags().StringVar(&urls, "urls", "./urls.txt", "simple .txt file with URL's to visit")
	root.Flags().BoolVarP(&verbose, "verbose", "v", false, "enables verbose mode")

	return root
}

func readURLS(r io.Reader) ([]string, error) {
	urls := []string{}
	input := bufio.NewScanner(r)
	for input.Scan() {
		url := input.Text()
		if url == "" {
			continue
		}

		idx := strings.Index(url, "://")
		if idx < 0 {
			url = "https://" + url // default https
		} else if idx == 0 {
			log.Printf("there is no valid scheme in url '%v'\n", url)
			continue
		}
		if !strings.HasPrefix(url, "http") {
			log.Printf("this is not supported scheme '%v'\n", url)
			continue
		}
		urls = append(urls, url)
	}

	return urls, input.Err()
}

func visit(site string, c *http.Client, agent string, delay time.Duration, verbose bool, debug bool, sema <-chan struct{}) {
	defer func() {
		time.Sleep(delay)
		<-sema
	}()

	code, err := request(c, site, agent)
	if err != nil {
		if debug {
			log.Printf("while making a request: %v", err)
		}

		return
	}

	if verbose {
		log.Println(site + " - " + code)
	}
}

func request(c *http.Client, url string, agent string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", agent)

	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}

	return resp.Status, nil
}

func randomDelay(delay time.Duration, randomDelay bool) time.Duration {
	if !randomDelay {
		return delay
	}

	r := rand.Intn(int(delay))
	return time.Duration(r)
}
