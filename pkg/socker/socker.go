// Copyright (c) 2018 China-HPC.

// Package socker implements a secure runner for docker containers.
package socker

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
)

// Socker provides a runner for docker.
type Socker struct {
}

// New would create a socker instance.
func New() *Socker {
	return &Socker{}
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
