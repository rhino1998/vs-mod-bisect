package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rhino1998/vs-mod-bisect/pkg/vsmod"
	"github.com/urfave/cli/v3"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	err := (&cli.Command{
		Name: "bisect",
		Action: func(ctx context.Context, c *cli.Command) error {
			path := c.Args().First()
			fmt.Printf("Loading mods from directory: %s\n", path)

			infos, err := vsmod.ReadModInfos(os.DirFS(path), ".")
			if err != nil {
				return err
			}

			components, err := vsmod.SortedComponents(infos)
			if err != nil {
				return err
			}

			var readd bool
			for {
				if len(components) == 1 {
					fmt.Print("Found bug in component:\n")
					printComponent(components[0])
				}

				left, right, err := vsmod.BisectComponents(components)
				if err != nil {
					return err
				}

				if readd {
					fmt.Printf("Re-Add:\n")
					for _, comp := range right {
						fmt.Print("--")
						printComponent(comp)
					}
				} else {
					fmt.Printf("Remove:\n")
					for _, comp := range left {
						fmt.Print("--")
						printComponent(comp)
					}
				}

				fmt.Printf("Bug still present? ")
				var resp string
				fmt.Scanf("%s", &resp) // Wait for user input
				if resp == "y" || resp == "yes" {
					components = right
					readd = false
				} else {
					components = left
				}
			}
		},
	}).Run(ctx, os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}

func printComponent(component []*vsmod.InfoWithFilename) {
	for _, info := range component {
		fmt.Printf("- %s (%s)\n", info.FileName, info.Name)
	}
}
