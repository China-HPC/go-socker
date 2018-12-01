#!/bin/bash

## You should configure Slurm to enable epilog. This script will be excuted
## after each Slurm job termnated to delete corresponding container.
recordFile=/var/lib/socker/epilog/$SLURM_JOB_ID
if [ -f $recordFile ];then
    echo "clean docker container for job: $SLURM_JOB_ID"
    containerName=`cat $recordFile`
    ownerRecord=/var/lib/socker/epilog/$containerName
    pidRecord=$ownerRecord"-pids"
    docker rm -f $containerName
    for pid in `cat $pidRecord`; do
        kill -- $pid
    done
    rm -f $recordFile $ownerRecord $pidRecord
fi
