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
- Docker 18.06+
- You MUST have a group docker and a user dockerroot who is member of ONLY the docker group. The docker run command will be executed as dockerroot.
- You SHOULD enable Linux namespaces and Docker `userns-remap` feature to make `socker` safer. Read the [Docker document](https://docs.docker.com/engine/security/userns-remap/) to know more about `userns-remap` please.
- Golang 1.6+(For development ONLY)

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

You should run command to sync `Docker` images to `socker`, simple usage:

```bash
socker images sync
```

#### sync filter

- Docker filter: use `--filter` or `-f` flag to specify filter to sync Docker filtered images，[read the document to know more](https://docs.docker.com/engine/reference/commandline/images/#filtering)
- Repository filter: use `--repo` or `-r` flag to specify repo filter to sync the images which contain specific keyword

```bash
## Example
## sync harbor.hpc.com/* images.
socker images sync --repo "harbor.hpc.com"

## sync docker filtered images.
socker images sync --filter "reference=ubuntu*"
```

#### customized images

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

## Security

Socker should work with Docker daemon which `userns-remap` feature has enbaled.

> The best way to prevent privilege-escalation attacks from within a container is to configure your container’s applications to run as unprivileged users, For containers whose processes must run as the root user within the container, you can re-map this user to a less-privileged user on the Docker host.

`socker` will default mount a swap directory(`$HOME/container`) to container, the root user of container can write data into this safe directory with `userns-remap` specified user's permission.

You can also use `socker` with Docker daemon without `userns-remap`, but this is dangerous. Safe or convenient, you can only choose one of them at present.

## Support and Bug Reports

## License

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FChina-HPC%2Fgo-socker.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2FChina-HPC%2Fgo-socker?ref=badge_large)
