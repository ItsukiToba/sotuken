#!/bin/bash -e

mkdir -p "/var/lib/docker/overlay2/neoimage/work-test-$2"
mkdir "/var/lib/docker/overlay2/neoimage/upper-test-$2"
mkdir "/var/lib/docker/overlay2/neoimage/merge-test-$2"

mount $1
