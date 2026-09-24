package main

import (
	"fmt"
	"os"

	"git.golder.lan/rossgolderltd/vault-tool/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
