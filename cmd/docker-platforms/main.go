package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/wantedly/container-platform-tools/dockerplatforms"
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "docker-platforms image...",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			resolver := dockerplatforms.NewImageTools()
			cache, err := dockerplatforms.NewYAMLCache(ctx, "image-platforms.yaml")
			if err != nil {
				return errors.Wrap(err, "initializing cache")
			}
			defer cache.WriteBack(ctx)
			inspector := dockerplatforms.New(resolver, cache)
			for _, image := range args {
				platforms, err := inspector.GetPlatforms(ctx, image)
				if err != nil {
					return errors.Wrap(err, "Inspecting image platform")
				}
				fmt.Printf("Platforms for %s: %v\n", image, platforms)
			}
			return nil
		},
	}

	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
