// Copyright (c) 2018 China-HPC.

// Package socker implements a secure runner for docker containers.
package socker

import "fmt"

// Socker provides a runner for docker.
type Socker struct {
}

// New would create a socker instance.
func New() *Socker {
	return &Socker{}
}

// ListImages would list all available images from your images hub.
func (s *Socker) ListImages() {
	fmt.Println("all images")
}

// RunImage would run a container by regular user.
func (s *Socker) RunImage(image string, command []string) {
	fmt.Println(image, command)
}
