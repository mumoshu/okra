package main

import (
	"fmt"
	"os"

	"github.com/mumoshu/okra/pkg/okractl/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		if os.Getenv("TRACE") != "" {
			fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}
