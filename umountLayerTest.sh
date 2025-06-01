#!/bin/bash -e
umount $1

rm -rf "/var/lib/docker/overlay2/neoimage/work-test-$2"
rm -rf "/var/lib/docker/overlay2/neoimage/upper-test-$2"
rm -rf "/var/lib/docker/overlay2/neoimage/merge-test-$2"