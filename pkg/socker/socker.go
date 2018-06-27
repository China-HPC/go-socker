// Copyright (c) 2018 China-HPC.

// Package socker implements a secure runner for docker containers.
package socker

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"

	log "github.com/Sirupsen/logrus"
	flags "github.com/jessevdk/go-flags"
	"github.com/kr/pty"
	"github.com/urfave/cli"
)

const (
	cmdDocker = "docker"
)

// Socker provides a runner for docker.
type Socker struct {
	DockerUID    string
	DockerGID    string
	CurrentUID   string
	CurrentUser  string
	CurrentGID   string
	CurrentGroup string
	HomeDir      string
}

// Opts represents the socker supported options.
type Opts struct {
	Volumes     []string `short:"v" long:"volume"`
	TTY         bool     `short:"t" long:"tty"`
	Interactive bool     `short:"i" long:"interactive"`
}

// New creates a socker instance.
func New(verbose bool) (*Socker, error) {
	if verbose {
		log.SetLevel(log.DebugLevel)
	}
	log.SetOutput(os.Stdout)
	s := &Socker{}
	err := s.checkPrerequisite()
	if err != nil {
		return nil, err
	}
	return s, nil
}

// ListImages lists all available images from registry.
func (s *Socker) ListImages(config string) error {
	info, err := os.Stat(config)
	if err != nil {
		log.Error(err)
		return err
	}
	if info.IsDir() {
		files, err := ioutil.ReadDir(config)
		if err != nil {
			log.Error(err)
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
			log.Error(err)
			return err
		}
		fmt.Println(string(data))
	}
	return nil
}

// RunImage runs container.
func (s *Socker) RunImage(command []string) error {
	args := []string{"run",
		"-v", fmt.Sprintf("%s:%s", s.HomeDir, s.HomeDir),
	}
	opts := Opts{}
	_, err := flags.ParseArgs(&opts, command)
	if err != nil {
		log.Error("parse command args failed: %v", err)
		return err
	}
	if !s.isVolumePermit(opts.Volumes) {
		return fmt.Errorf("illegal volume mount")
	}
	args = append(args, command...)
	log.Debug("docker run args: %v", args)
	cmd := exec.Command(cmdDocker, args...)
	if opts.TTY {
		return runWithPty(cmd)
	}
	return cmd.Run()
}

func (s *Socker) isVolumePermit(vols []string) bool {
	// TODO(xhzhang): check volumes permission.
	return true
}

func runWithPty(cmd *exec.Cmd) error {
	tty, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("docker command exec failed: %v", err)
	}
	go func() {
		io.Copy(os.Stdout, tty)
	}()
	io.Copy(tty, os.Stdin)
	return nil
}

func (s *Socker) checkPrerequisite() error {
	if !isCommandAvailable("docker") {
		return cli.NewExitError("docker command not found, make sure Docker is installed...", 127)
	}
	u, err := user.Lookup("dockerroot")
	if err != nil {
		return cli.NewExitError("there must exist a user 'dockerroot' and a group 'docker'", 1)
	}
	s.DockerUID = u.Uid
	g, err := user.LookupGroup("docker")
	if err != nil {
		return cli.NewExitError("there must exist a user 'dockerroot' and a group 'docker'", 1)
	}
	s.DockerGID = g.Gid
	gids, err := u.GroupIds()
	if err != nil && isMemberOfGroup(gids, u.Gid) {
		return cli.NewExitError("the user 'dockerroot' must be a member of the 'docker' group", 2)
	}
	current, err := user.Current()
	if err != nil {
		return cli.NewExitError("can't get current user info", 2)
	}
	s.CurrentUID = current.Uid
	s.CurrentUser = current.Username
	s.CurrentGID = current.Gid
	currentGroup, err := user.LookupGroupId(s.CurrentGID)
	if err != nil {
		return cli.NewExitError("can't get current user's group info", 2)
	}
	s.CurrentGroup = currentGroup.Name
	s.HomeDir = current.HomeDir
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
