#!/bin/bash -e

mkdir "/var/lib/docker/overlay2/neoimage/work-$2"
mkdir "/var/lib/docker/overlay2/neoimage/upper-$2"
mkdir -p $3
mount $1
