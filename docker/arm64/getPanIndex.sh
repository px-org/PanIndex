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
curl -sOL "https://github.com/libsgh/PanIndex/releases/download/${version}/PanIndex-${version}-linux-arm64.tar.gz"
sha256sum "PanIndex-"${version}"-linux-arm64.tar.gz"
tar -xvzf "PanIndex-"${version}"-linux-arm64.tar.gz"
rm -rf README.md LICENSE
chmod +x PanIndex
