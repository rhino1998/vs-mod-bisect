package main

import (
	"cmp"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
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

			modPath := filepath.Join(path, "Mods")
			disablePath := filepath.Join(path, "DisabledMods")
			err := os.Mkdir(disablePath, 0755)
			if err != nil {
				return err
			}

			infos, err := vsmod.ReadModInfos(os.DirFS(modPath), ".")
			if err != nil {
				return err
			}

			components, err := vsmod.SortedComponents(infos)
			if err != nil {
				return err
			}

			var readd bool
			for {
				if len(components) == 0 {
					fmt.Print("No components found, exiting.\n")
					return nil
				}
				if len(components) == 1 {
					fmt.Print("Found bug in component:\n")
					printComponent(components[0])
					return nil
				}

				left, right, err := vsmod.BisectComponents(components)
				if err != nil {
					return err
				}

				if readd {
					fmt.Printf("Enable:\n")
					printComponentsSorted(right)
					err := moveMods(disablePath, modPath, right)
					if err != nil {
						return fmt.Errorf("failed to enable mods: %w", err)
					}
				} else {
					fmt.Printf("Disable:\n")
					printComponentsSorted(left)
					err := moveMods(modPath, disablePath, left)
					if err != nil {
						return fmt.Errorf("failed to disable mods: %w", err)
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
					readd = true
					fmt.Printf("Disable:\n")
					printComponentsSorted(right)
					err := moveMods(modPath, disablePath, right)
					if err != nil {
						return fmt.Errorf("failed to disable mods: %w", err)
					}
				}
			}
		},
	}).Run(ctx, os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}

func moveMods(from, to string, components [][]*vsmod.InfoWithFilename) error {
	for _, component := range components {
		for _, info := range component {
			fromMod := filepath.Join(from, info.FileName)
			toMod := filepath.Join(to, info.FileName)
			_ = os.Remove(toMod) // Remove the mod if it exists, to avoid conflicts
			if err := os.Rename(fromMod, toMod); err != nil {
				return fmt.Errorf("failed to enable mod %s: %w", info.FileName, err)
			}
		}
	}

	return nil
}

func printComponent(component []*vsmod.InfoWithFilename) {
	fmt.Print("--\n")
	for _, info := range component {
		fmt.Printf("- %s (%s)\n", info.FileName, info.Name)
	}
}

func printComponentsSorted(components [][]*vsmod.InfoWithFilename) {
	var s []*vsmod.InfoWithFilename
	for _, component := range components {
		for _, info := range component {
			s = append(s, info)
		}
	}
	slices.SortFunc(s, func(a, b *vsmod.InfoWithFilename) int {
		return cmp.Compare(strings.ToLower(a.FileName), strings.ToLower(b.FileName))
	})

	for _, info := range s {
		fmt.Printf("- %s (%s)\n", info.FileName, info.Name)
	}
}
