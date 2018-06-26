// Copyright (c) 2018 China-HPC.

package main

import (
	"log"
	"os"
	"os/exec"
	"os/user"

	"github.com/China-HPC/go-socker/pkg/socker"
	"github.com/urfave/cli"
)

var (
	verbose   bool
	dockeruid string
	dockergid string
)

func main() {
	s := socker.New()

	app := cli.NewApp()
	app.Name = "socker"
	app.Usage = "Secure runner for Docker containers"
	app.Version = "0.1.0"
	app.Before = checkPrerequisite
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "verbose",
			Destination: &verbose,
			Usage:       "run in verbose mode",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "images",
			Usage: "List available images that found in your images hub",
			Action: func(c *cli.Context) {
				s.ListImages()
			},
		},
		{
			Name:  "run",
			Usage: "run a container from IMAGE executing COMMAND as regular user",
			Action: func(c *cli.Context) {
				image := c.Args().First()
				command := c.Args().Tail()
				s.RunImage(image, command)
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func checkPrerequisite(ctx *cli.Context) error {
	if !isCommandAvailable("docker") {
		return cli.NewExitError("docker command not found, make sure Docker is installed...", 127)
	}
	u, err := user.Lookup("dockerroot")
	if err != nil {
		return cli.NewExitError("There must exist a user 'dockerroot' and a group 'docker'", 1)
	}
	dockeruid = u.Uid
	g, err := user.LookupGroup("docker")
	if err != nil {
		return cli.NewExitError("There must exist a user 'dockerroot' and a group 'docker'", 1)
	}
	dockergid = g.Gid
	gids, err := u.GroupIds()
	if err != nil && isMemberOfGroup(gids, u.Gid) {
		return cli.NewExitError("The user 'dockerroot' must be a member of the 'docker' group", 2)
	}
	return nil
}

func isMemberOfGroup(gids []string, gid string) bool {
	for _, id := range gids {
		if id == gid {
			return true
		}
	}
	return false
}

func isCommandAvailable(name string) bool {
	cmd := exec.Command("command", "-v", name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
