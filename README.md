# go-socker
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FChina-HPC%2Fgo-socker.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2FChina-HPC%2Fgo-socker?ref=badge_shield)


A wrapper for secure running of Docker containers on Slurm implement in Golang. Inspired by the paper _[Enabling Docker Containers for High-Performance and Many-Task Computing](https://ieeexplore.ieee.org/document/7923813/)_ and [socker](https://github.com/unioslo/socker).

## Introduction

Socker is secure for enabling unprivileged users to run Docker containers. It mainly does two things:

- It enforces running containers within as the user not as root
- When it is called inside a Slurm job, it enforces the inclusion of containers in the [cgroups assigned by Slurm to the parent jobs](https://slurm.schedmd.com/cgroups.html)

## Prerequisite

### MUST

- CentOS/Redhat and Debian have been tested
- Docker 1.6+
- Golang 1.6+(if you want to build from source)
- You MUST have a group docker and a user dockerroot who is member of ONLY the docker group. The docker run command will be executed as dockerroot.

To add the dockerroot user to docker group:

```bash
usermod -aG docker dockerroot
```

### Optional

- Slurm is not a prerequisite, but if you run socker inside a Slurm job, it will put the container under Slurm's control.
- `libcgroup-tools` should be installed for cgroup limit set.

## Installation

### Build from source

make sure you have installed [`go`](https://golang.org/dl/) and [`glide`](https://github.com/Masterminds/glide), then:

```bash
make install
```

### Configure images

You should run command to sync `Docker` images to `socker`:

```bash
socker images sync
```

Or define your images config in `/var/lib/socker/images.yaml` file manually before using `socker images` command.

### Configure with slurm (Optional)

If you want to delete containers after Slurm job terminated, you should use the `epilog.sh` script in scripts directory as Slurm epilog script.

## Quick Start

Use socker just like docker, for example:

```bash
socker run -it ubuntu bash
```

Run socker --help to know more:

```txt
NAME:
   socker - Secure runner for Docker containers

USAGE:
   socker [global options] command [command options] [arguments...]

VERSION:
   0.1.0

COMMANDS:
     images   List images that defined in image.yaml file or sync images from Docker to socker.
     run      run a container from IMAGE executing COMMAND as regular user
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --verbose      run in verbose mode
   --epilog       run with Slurm epilog enabled
   --help, -h     show help
   --version, -v  print the version
```

## Support and Bug Reports

## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FChina-HPC%2Fgo-socker.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2FChina-HPC%2Fgo-socker?ref=badge_large)
