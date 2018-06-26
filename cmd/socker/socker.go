// Copyright (c) 2018 China-HPC.

package main

import (
	"log"
	"os"

	"github.com/China-HPC/go-socker/pkg/socker"
	"github.com/urfave/cli"
)

func main() {
	s := socker.New()

	app := cli.NewApp()
	app.Name = "socker"
	app.Usage = "Secure runner for Docker containers"
	app.Version = "0.1.0"
	app.Commands = []cli.Command{
		{
			Name:  "images",
			Usage: "List available images that found in your images registry from `FILE` or `PATH`",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config, c",
					Usage: "images registry from `FILE` or `PATH`",
				},
			},
			Action: func(c *cli.Context) error {
				if c.String("config") == "" {
					return cli.NewExitError("Need images registry's FILE or PATH", 1)
				}
				err := s.ListImages(c.String("config"))
				if err != nil {
					return cli.NewExitError(err.Error(), 1)
				}
				return nil
			},
		},
		{
			Name:  "run",
			Usage: "run a container from IMAGE executing COMMAND as regular user",
			Action: func(c *cli.Context) {
				image := c.Args().First()
				command := c.Args().Tail()
				s.RunImage(image, command)
				return
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
