// Copyright (c) 2018 China-HPC.

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/China-HPC/go-socker/pkg/socker"
	"github.com/urfave/cli"
)

var (
	verbose       bool
	epilogEnabled bool
	insecure      bool
	ptyRows       int
	ptyCols       int
	s             *socker.Socker
)

func main() {
	app := cli.NewApp()
	app.Name = "socker"
	app.Usage = "Secure runner for Docker containers"
	app.Version = "0.1.0"
	app.Before = appInit
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "verbose",
			Destination: &verbose,
			Usage:       "run in verbose mode",
		},
		cli.BoolFlag{
			Name:        "epilog",
			Destination: &epilogEnabled,
			Usage:       "run with Slurm epilog enabled",
		},
		cli.BoolFlag{
			Name:        "insecure",
			Destination: &insecure,
			Usage:       "run in insecure mode, strongly not recommended",
		},
		cli.IntFlag{
			Name:        "rows",
			Destination: &ptyRows,
			Usage:       "the rows value of pty size",
			Value:       35,
		},
		cli.IntFlag{
			Name:        "cols",
			Destination: &ptyCols,
			Usage:       "the cols value of pty size",
			Value:       87,
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "images",
			Usage: "List images that defined in image.yaml file or sync images from Docker to socker.",
			Subcommands: []cli.Command{
				{
					Name:  "list",
					Usage: "list all images",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "config, c",
							Usage: "images config file",
						},
					},
					Action: func(c *cli.Context) error {
						err := s.PrintImages(c.String("config"))
						if err != nil {
							return cli.NewExitError(err.Error(), 1)
						}
						return nil
					},
				},
				{
					Name:  "sync",
					Usage: "sync images from docker (NOTE:common user have no permission to do this operation)",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "config, c",
							Usage: "images config file",
						},
						cli.StringFlag{
							Name:  "repo, r",
							Usage: "filter the repository",
						},
						cli.StringFlag{
							Name:  "filter, f",
							Usage: "filter the images",
						},
					},
					Before: func(c *cli.Context) error {
						if s.CurrentUID != "0" {
							log.Fatal("You have no permission to do this.")
						}
						return nil
					},
					Action: func(c *cli.Context) error {
						err := s.SyncImages(c.String("config"),
							c.String("repo"), c.String("filter"))
						if err != nil {
							return cli.NewExitError(err.Error(), 1)
						}
						return nil
					},
				},
			},
		},
		{
			Name:            "run",
			Usage:           "run a container from IMAGE executing COMMAND as regular user",
			SkipFlagParsing: true,
			Action: func(c *cli.Context) error {
				err := s.RunImage(c.Args())
				if err != nil {
					return cli.NewExitError(err, 1)
				}
				return nil
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func appInit(ctx *cli.Context) error {
	var err error
	conf := &socker.Config{
		Verbose:       verbose,
		EpilogEnabled: epilogEnabled,
		Insecure:      insecure,
		PtyCols:       ptyCols,
		PtyRows:       ptyRows,
	}
	s, err = socker.New(conf)
	if err != nil {
		log.Fatal(fmt.Sprintf("init socker failed: %v", err))
		os.Exit(2)
	}
	return nil
}
