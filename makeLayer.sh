#!/bin/bash -e
while IFS= read -r line
do
    if [ -L "/var/lib/docker/overlay2/neoimage/merge-test-$2$line" ]; then
        dir_name=$(dirname "$line")
        if [ ! -d "/var/lib/docker/overlay2/neoimage/$1/$dir_name" ]; then
            mkdir -p /var/lib/docker/overlay2/neoimage/$1/$dir_name
        fi
        cp -P "/var/lib/docker/overlay2/neoimage/merge-test-$2$line" "/var/lib/docker/overlay2/neoimage/$1/$dir_name"
    elif [ -d "/var/lib/docker/overlay2/neoimage/merge-test-$2$line" ]; then
        mkdir -p /var/lib/docker/overlay2/neoimage/$1/$line
        permissions=$(stat -c "%a" "/var/lib/docker/overlay2/neoimage/merge-test-$2$line")
        chmod $permissions /var/lib/docker/overlay2/neoimage/$1/$line
    elif [ -c "/var/lib/docker/overlay2/neoimage/merge-test-$2$line" ]; then
        dir_name=$(dirname "$line")
        if [ ! -d "/var/lib/docker/overlay2/neoimage/$1/$dir_name" ]; then
            mkdir -p /var/lib/docker/overlay2/neoimage/$1/$dir_name
        fi
        MAJOR=$(ls -l "/var/lib/docker/overlay2/neoimage/merge-test-$2$line" | awk '{print $5}' | tr ',' ' ')
        MINOR=$(ls -l "/var/lib/docker/overlay2/neoimage/merge-test-$2$line" | awk '{print $6}')
        mknod "/var/lib/docker/overlay2/neoimage/$1$line" c $MAJOR $MINOR
        chmod 666 "/var/lib/docker/overlay2/neoimage/$1$line"
    elif [ -p "/var/lib/docker/overlay2/neoimage/merge-test-$2$line" ]; then
        dir_name=$(dirname "$line")
        if [ ! -d "/var/lib/docker/overlay2/neoimage/$1/$dir_name" ]; then
            mkdir -p /var/lib/docker/overlay2/neoimage/$1/$dir_name
        fi
        mkfifo "/var/lib/docker/overlay2/neoimage/$1$line"
    else
        dir_name=$(dirname "$line")
        if [ ! -d "/var/lib/docker/overlay2/neoimage/$1/$dir_name" ]; then
            mkdir -p /var/lib/docker/overlay2/neoimage/$1/$dir_name
        fi
        cp -P "/var/lib/docker/overlay2/neoimage/merge-test-$2$line" "/var/lib/docker/overlay2/neoimage/$1/$dir_name"
    fi
done < "/go/src/github.com/docker/docker/neoimage/path.dat"
rm /go/src/github.com/docker/docker/neoimage/path.dat
