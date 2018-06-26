# go-socker

A wrapper for secure running of Docker containers on Slurm implement in Golang. Inspired by the paper _[Enabling Docker Containers for High-Performance and Many-Task Computing](https://ieeexplore.ieee.org/document/7923813/)_ and [socker](https://github.com/unioslo/socker).

## Introduction

Socker is secure for enabling unprivileged users to run Docker containers. It mainly does two things:

- It enforces running containers within as the user not as root
- When it is called inside a Slurm job, it enforces the inclusion of containers in the [cgroups assigned by Slurm to the parent jobs](https://slurm.schedmd.com/cgroups.html)

## Quick Start

## Prerequisite

### MUST

- CentOS/Redhat at present
- Docker 1.6+
- Golang 1.8.3+
- You MUST have a group docker and a user dockerroot who is member of ONLY the docker group. The docker run command will be executed as dockerroot.

To add the dockerroot user to docker group:

```bash
usermod -aG docker dockerroot
```

### Optional

- Slurm is not a prerequisite, but if you run socker inside a Slurm job, it will put the container under Slurm's control.
- `libcgroup-tools` should be installed for cgroup limit set.

## Support and Bug Reports