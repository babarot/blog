package main

import (
	"fmt"
	"os"

	"github.com/babarot/blog/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "blog: %v\n", err)
		os.Exit(1)
	}
}
