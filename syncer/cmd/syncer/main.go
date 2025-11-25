package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nir414/pc-setup/syncer/internal/app"
)

func main() {
	ctx := context.Background()
	application := app.New()
	if err := application.Run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "syncer: %v\n", err)
		os.Exit(1)
	}
}
