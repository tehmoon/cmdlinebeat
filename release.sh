#!/bin/sh

#set -e
#
#libbeat_version=6.5
#internal_version=6.5.4
#
#image_name=cmdlinebeat_release
#
#docker rm -f "${image_name}" || true
#docker run -d --name "${image_name}" --rm --entrypoint /bin/sleep golang:rc-alpine 1000000000
#
#docker cp . "${image_name}:/tmp/cmdlinebeat"
#
#docker exec "${image_name}" chown -R nobody:nogroup /tmp/cmdlinebeat
#docker exec "${image_name}" apk add --no-cache git musl-dev gcc
#docker exec -e HOME=/tmp/cmdlinebeat -e libbeat_version="${libbeat_version}" -e internal_version="${internal_version}" -u nobody "${image_name}" /bin/sh -c "cd && sh ./build.sh"
#rm -rf .builds || true
#docker cp "${image_name}":/tmp/cmdlinebeat/.builds .
