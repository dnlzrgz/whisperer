package main

import "github.com/danielkvist/whisperer/cmd"

func main() {
	root := cmd.Root()
	root.Execute()
}
