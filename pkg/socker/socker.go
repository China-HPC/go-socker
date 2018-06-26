// Copyright (c) 2018 China-HPC.

// Package socker implements a secure runner for docker containers.
package socker

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"

	"github.com/urfave/cli"
)

// Socker provides a runner for docker.
type Socker struct {
	DockerUID string
	DockerGID string
}

// New would create a socker instance.
func New() (*Socker, error) {
	s := &Socker{}
	err := s.checkPrerequisite()
	if err != nil {
		return nil, err
	}
	return s, nil
}

// ListImages would list all available images from your images registry.
func (s *Socker) ListImages(config string) error {
	info, err := os.Stat(config)
	if err != nil {
		log.Fatal(err)
		return err
	}
	if info.IsDir() {
		files, err := ioutil.ReadDir(config)
		if err != nil {
			log.Fatal(err)
			return err
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			data, err := ioutil.ReadFile(path.Join(config, file.Name()))
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		}
	} else {
		data, err := ioutil.ReadFile(config)
		if err != nil {
			log.Fatal(err)
			return err
		}
		fmt.Println(string(data))
	}
	return nil
}

// RunImage would run a container by regular user.
func (s *Socker) RunImage(image string, command []string) {
	fmt.Println(image, command)
}

func (s *Socker) checkPrerequisite() error {
	if !isCommandAvailable("docker") {
		return cli.NewExitError("docker command not found, make sure Docker is installed...", 127)
	}
	u, err := user.Lookup("dockerroot")
	if err != nil {
		return cli.NewExitError("There must exist a user 'dockerroot' and a group 'docker'", 1)
	}
	s.DockerUID = u.Uid
	g, err := user.LookupGroup("docker")
	if err != nil {
		return cli.NewExitError("There must exist a user 'dockerroot' and a group 'docker'", 1)
	}
	s.DockerGID = g.Gid
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
