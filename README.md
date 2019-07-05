# Whisperer

Whisperer is a simple Go program that makes HTTP request constantly in order to generate random HTTP/DNS traffic noise.

> Whisperer is a project inspired by [Noisy](https://github.com/1tayH/noisy).

## Install

### Go

```bash
go get github.com/danielkvist/whisperer
```

### Cloning the repository

```bash
# First, clone the repository
git clone https://github.com/danielkvist/whisperer

# Then navigate into the whisperer directory
cd whisperer

# Run
go run main.go
```

## Options

Whisperer can accept a number of command line arguments:

```text
$ whisperer --help
whisperer makes HTTP request constantly in order to generate random HTTP/DNS traffic noise.

Usage:
  whisperer [flags]

Flags:
      --agent string        user agent (default "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:67.0) Gecko/20100101 Firefox/67.0")
      --goroutines string   number of goroutines (default "1")
  -h, --help                help for whisperer
      --urls string         simple .txt file with URL's to visit (default "./urls.txt")

```

## URLs file

This file is from which Whisperer will extract the different URLs that will be visiting.

> You can see an example of how this file should be [here](https://github.com/danielkvist/whisperer/blob/master/urls.txt).

## Things to implement/improve

* Docker image.
* Tests.
* Command line arguments.

If you know about anything else I can improve or add please, don't hesitate to let me know!
