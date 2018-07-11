#!/bin/sh
version=6.3.1

rm *.tar.gz 2>/dev/null

echo Releasing for version ${version}

for os in linux openbsd darwin
do
  GOOS=${os}
  GOARCH=amd64
  dir=cmdlinebeat-${version}-${GOOS}-x86_64

  GOOS=${GOOS} GOARCH=${GOARCH} go build
  mkdir ${dir}
  cp -r README.md cmdlinebeat cmdlinebeat.yml ${dir}
  tar -cvvzf ${dir}.tar.gz ${dir}
  rm -rf ./${dir}
done
