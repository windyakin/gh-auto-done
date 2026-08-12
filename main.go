package main

import (
	"os"

	"git.pepabo.com/windyakin/gh-auto-done/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
