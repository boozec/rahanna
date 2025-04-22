#!/bin/sh

if [ $# -eq 0 ]; then
    echo "You must pass the version number."
    exit 1
fi

sed -i "15s/.*/LABEL version=\"$1\"/" docker/api/Dockerfile
sed -i "17s/.*/LABEL version=\"$1\"/" docker/ui/Dockerfile
git add docker
git commit -m "release: version $1"
git tag -a "$1" -m "Version $1"
