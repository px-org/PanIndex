#!/bin/bash
version=$PAN_INDEX_VERSION
echo $PAN_INDEX_VERSION
version=""
if [ "$version" = "" ]
then
    version=`curl --silent "https://api.github.com/repos/libsgh/PanIndex/releases/latest" \
        | grep '"tag_name":' \
        | sed -E 's/.*"([^"]+)".*/\1/'`
fi
docker build -t "iicm/pan-index:amd64-${version}" .
cd arm64
docker build -t "iicm/pan-index:arm64-${version}" .
docker push "iicm/pan-index:amd64-${version}"
docker push "iicm/pan-index:arm64-${version}"
docker manifest create "iicm/pan-index:${version}" "iicm/pan-index:amd64-${version}" "iicm/pan-index:arm64-${version}" --amend
docker manifest inspect "iicm/pan-index:${version}" --verbose
docker manifest push "iicm/pan-index:${version}" --purge
docker manifest create "iicm/pan-index:latest" "iicm/pan-index:amd64-${version}" "iicm/pan-index:arm64-${version}" --amend
docker manifest inspect "iicm/pan-index:latest" --verbose
docker manifest push "iicm/pan-index:latest" --purge