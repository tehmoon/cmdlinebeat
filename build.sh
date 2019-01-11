#!/bin/sh

set -e

[ "x${libbeat_version}" = "x" ] && libbeat_version=6.5
[ "x${internal_version}" = "x" ] && internal_version=6.5.4

rm -rf .go .builds || true
mkdir .go .builds || true

rm *.tar.gz 2>/dev/null || true

export GOPATH="$(pwd)/.go"

go get -v ./... || true

cd "${GOPATH}/src/github.com/elastic/beats"
git checkout -b "${libbeat_version}" "origin/${libbeat_version}"
cd -

echo Releasing for internal_version "${internal_version}"

for os in linux openbsd darwin
do
  GOOS="${os}"
  GOARCH=amd64
  dir="cmdlinebeat-${internal_version}-${GOOS}-x86_64"

  GOOS="${GOOS}" GOARCH="${GOARCH}" go build -tags netgo -ldflags '-w -s'
  mkdir "${dir}"
  cp -r README.md cmdlinebeat cmdlinebeat.yml "${dir}"
  tar -cvvzf ".builds/${dir}.tar.gz" "${dir}"
  rm -rf "./${dir}"
done
