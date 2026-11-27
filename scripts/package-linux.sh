#!/usr/bin/env bash
set -euo pipefail
cd -- "$(dirname -- "$0")/.."

[[ $(go env GOOS) == linux && $(go env GOARCH) == amd64 ]]
version=$(tr -d '\r\n' < VERSION)
[[ $version =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$ ]]
archive="dynamicbrowser-$version-linux-amd64.tar.gz"
stage=$(mktemp -d)
trap 'rm -rf -- "$stage"' EXIT
mkdir -p dist "$stage/package" "$stage/unpack"
go build -ldflags "-X main.version=$version" -o "$stage/package/dynamicbrowser" ./cmd/dynamicbrowser
cp -- config.example.yaml internal/config/config.schema.json README.md LICENSE "$stage/package/"
tar -czf "$stage/$archive" -C "$stage/package" dynamicbrowser config.example.yaml config.schema.json README.md LICENSE
(cd "$stage"; sha256sum "$archive" > "$archive.sha256"; sha256sum -c "$archive.sha256")
tar -xzf "$stage/$archive" -C "$stage/unpack"
[[ $("$stage/unpack/dynamicbrowser" --version) == "$version" ]]

export XDG_CONFIG_HOME="$stage/config" XDG_CACHE_HOME="$stage/cache" PROBE_RESULT="$stage/result"
mkdir -p "$XDG_CONFIG_HOME/dynamicbrowser"
printf '%s\n' '#!/bin/sh' 'printf "%s\n" "$1" > "$PROBE_RESULT"' > "$stage/unpack/probe"
chmod +x "$stage/unpack/probe"
printf '%s\n' 'default: probe' 'browsers:' '  probe: {exe: ./probe}' > "$XDG_CONFIG_HOME/dynamicbrowser/config.yaml"
url=https://example.test/package-smoke
"$stage/unpack/dynamicbrowser" "$url"
for _ in {1..50}; do
  [[ ! -s $PROBE_RESULT ]] || break
  sleep 0.02
done
grep -Fxq "$url" "$PROBE_RESULT"
mv -- "$stage/$archive" "$stage/$archive.sha256" dist/
