package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"github.com/rhino1998/vs-mod-bisect/pkg/vsmod"
	"github.com/urfave/cli/v3"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	(&cli.Command{
		Name: "bisect",
		Action: func(ctx context.Context, c *cli.Command) error {
			path := c.Args().First()
			fmt.Printf("Loading mods from directory: %s\n", path)

			infos, err := vsmod.ReadModInfos(os.DirFS(path), ".")
			if err != nil {
				return err
			}

			for {
				if len(infos) == 1 {
					for _, info := range infos {
						fmt.Printf("Found bug in mod: %s\n", info.Name)
					}
				}

				left, right, err := vsmod.Bisect(infos)
				if err != nil {
					return err
				}

				remove := make([]string, 0, len(left)+len(right))
				for path := range left {
					remove = append(remove, path)
				}
				slices.Sort(remove)

				fmt.Printf("Remove:\n")
				for _, path := range remove {
					fmt.Printf("- %s\n", path)
				}

				var resp string
				fmt.Scanf("Bug still present? %s", &resp) // Wait for user input
				if resp == "y" || resp == "yes" {
					infos = right
				} else {
					infos = left
				}
			}
		},
	}).Run(ctx, os.Args)
}
