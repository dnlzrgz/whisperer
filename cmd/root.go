package cmd

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	sites      map[string]struct{}
	goroutines string
	agent      string
	urls       string
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

		input := bufio.NewScanner(f)
		for input.Scan() {
			sites["https://"+input.Text()] = struct{}{}
		}

		n, err := strconv.Atoi(goroutines)
		if err != nil {
			return err
		}

		client := &http.Client{}
		sema := make(chan struct{}, n)
		for {
			for k := range sites {
				sema <- struct{}{}
				go func(site string) {
					defer func() { <-sema }()

					req, err := http.NewRequest(http.MethodGet, site, nil)
					if err != nil {
						log.Println(err)
					}
					req.Header.Set("User-Agent", agent)

					log.Printf("visiting: %q", site)
					_, err = client.Do(req)
					if err != nil {
						log.Println(err)
					}
				}(k)
			}
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	sites = make(map[string]struct{})
	rootCmd.PersistentFlags().StringVar(&goroutines, "goroutines", "1", "number of goroutines")
	rootCmd.PersistentFlags().StringVar(&agent, "agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:67.0) Gecko/20100101 Firefox/67.0", "user agent")
	rootCmd.PersistentFlags().StringVar(&urls, "urls", "./urls.txt", "simple .txt file with URL's to visit")
}
