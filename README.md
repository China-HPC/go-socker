# go-socker

A wrapper for secure running of Docker containers on Slurm implement in Golang. Inspired by the paper _[Enabling Docker Containers for High-Performance and Many-Task Computing](https://ieeexplore.ieee.org/document/7923813/)_ and [socker](https://github.com/unioslo/socker).

## Introduction

Socker is secure for enabling unprivileged users to run Docker containers. It mainly does two things:

- It enforces running containers within as the user not as root
- When it is called inside a Slurm job, it enforces the inclusion of containers in the [cgroups assigned by Slurm to the parent jobs](https://slurm.schedmd.com/cgroups.html)

## Quick Start

## Prerequisite

## Support and Bug Reports