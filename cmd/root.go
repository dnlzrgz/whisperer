package cmd

import (
	"bufio"
	"fmt"
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
	sites      []string
	goroutines string
	agent      string
	urls       string
	verbose    bool
)

var rootCmd = &cobra.Command{
	Use:   "whisperer",
	Short: "whisperer makes HTTP request constantly in order to generate random HTTP/DNS traffic noise.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := readFile(urls); err != nil {
			return err
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
			go func() {
				defer func() { <-sema }()
				i := r.Intn(len(sites) - 1)
				s := sites[i]

				req, err := http.NewRequest(http.MethodGet, s, nil)
				if err != nil {
					log.Println(err)
				}
				req.Header.Set("User-Agent", agent)

				if verbose {
					log.Printf("visiting: %q", s)
				}

				_, err = client.Do(req)
				if err != nil {
					log.Println(err)
				}
			}()
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&goroutines, "goroutines", "1", "number of goroutines")
	rootCmd.PersistentFlags().StringVar(&agent, "agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:67.0) Gecko/20100101 Firefox/67.0", "user agent")
	rootCmd.PersistentFlags().StringVar(&urls, "urls", "./urls.txt", "simple .txt file with URL's to visit")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enables verbose")
}

func readFile(route string) error {
	f, err := os.Open(route)
	if err != nil {
		return fmt.Errorf("while reading file %v: %v", urls, err)
	}
	defer f.Close()

	input := bufio.NewScanner(f)
	for input.Scan() {
		site := input.Text()
		if !strings.HasPrefix(site, "https://") {
			site = "https://" + site
		}
		sites = append(sites, site)
	}

	return nil
}
