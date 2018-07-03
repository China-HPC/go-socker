# go-socker

A wrapper for secure running of Docker containers on Slurm implement in Golang. Inspired by the paper _[Enabling Docker Containers for High-Performance and Many-Task Computing](https://ieeexplore.ieee.org/document/7923813/)_ and [socker](https://github.com/unioslo/socker).

## Introduction

Socker is secure for enabling unprivileged users to run Docker containers. It mainly does two things:

- It enforces running containers within as the user not as root
- When it is called inside a Slurm job, it enforces the inclusion of containers in the [cgroups assigned by Slurm to the parent jobs](https://slurm.schedmd.com/cgroups.html)

## Prerequisite

### MUST

- CentOS/Redhat at present
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

make sure you have installed `go` and `glide`, then:

```bash
make install
```

### Download released

1. Download from release list
2. Run those commands to install:

```bash
chown root:root socker
chmod +xs socker
mv socker /usr/bin/
```

### Configure images

You should define your images config in `/var/lib/socker/images.yaml` file manually before using `socker images` command.

### Configure with slurm (Optional)

If you want to delete containers after Slurm job terminated, you should use the `epilog.sh` script in scripts directory as Slurm epilog script.

## Quick Start

Use socker just like docker, for example:

```bash
socker run -it ubuntu bash
```

Run socker --help to know more:

```bash
NAME:
   socker - Secure runner for Docker containers

USAGE:
   socker [global options] command [command options] [arguments...]

VERSION:
   0.1.0

COMMANDS:
     images   List available images that found in your images registry from `FILE` or `PATH`
     run      run a container from IMAGE executing COMMAND as regular user
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --verbose      run in verbose mode
   --help, -h     show help
   --version, -v  print the version
```

## Support and Bug Reports