package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	goroutines string
	agent      string
	urls       string
	verbose    bool
)

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

		n, err := strconv.Atoi(goroutines)
		if err != nil {
			return err
		}

		client := &http.Client{}
		sema := make(chan struct{}, n)
		seed := rand.NewSource(time.Now().Unix())
		r := rand.New(seed)
		for {
			sema <- struct{}{}
			i := r.Intn(len(sites) - 1)
			s := sites[i]

			go request(sema, client, s, verbose)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&goroutines, "goroutines", "g", "1", "number of goroutines")
	rootCmd.PersistentFlags().StringVarP(&agent, "agent", "a", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:67.0) Gecko/20100101 Firefox/67.0", "user agent")
	rootCmd.PersistentFlags().StringVar(&urls, "urls", "./urls.txt", "simple .txt file with URL's to visit")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enables verbose mode")
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

func request(tokens <-chan struct{}, c *http.Client, url string, v bool) {
	defer func() { <-tokens }()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println(err)
	}
	req.Header.Set("User-Agent", agent)

	if v {
		log.Printf("visiting: %q", url)
	}

	_, err = c.Do(req)
	if err != nil {
		log.Println(err)
	}
}
