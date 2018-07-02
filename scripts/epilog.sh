#!/bin/bash

## You should configure Slurm to enable epilog. This script will be excuted 
## after each Slurm job termnated to delete corresponding container.

recordFile=/var/lib/socker/epilog/$SLURM_JOB_ID
containerName=`cat $recordFile`
docker rm -f $containerName
rm -f $recordFile