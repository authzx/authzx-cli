package main

import (
	"fmt"
	"os"

	"github.com/authzx/authzx-cli/internal/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "azx: %s\n", err)
		os.Exit(1)
	}
}
