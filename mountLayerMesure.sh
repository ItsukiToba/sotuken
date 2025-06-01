#!/bin/bash -e

if [ ! "$4" -eq 0 ]; then
    for i in $(seq 1 $4)
    do
        mkdir "/var/lib/docker/overlay2/neoimage/$i"
        for j in {1..10}; do
            truncate -s 1M /var/lib/docker/overlay2/neoimage/$i/file_$j
        done
    done
fi
mkdir "/var/lib/docker/overlay2/neoimage/work-$2"
mkdir "/var/lib/docker/overlay2/neoimage/upper-$2"
mkdir -p $3
mount $1
