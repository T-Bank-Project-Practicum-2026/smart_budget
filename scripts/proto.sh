#!/bin/sh
set -eu
mode=${1:-generate}
output=$(mktemp -d)
trap 'rm -rf "$output"' EXIT
cd /workspace/proto
buf lint
buf format --diff --exit-code
buf generate --template buf.gen.yaml --output "$output"
target=/workspace/backend/internal/gen
if [ "$mode" = check ]; then
  diff -ru "$target" "$output" || {
    echo 'Protobuf-код устарел. Запустите make proto.' >&2
    exit 1
  }
else
  mkdir -p "$target"
  # This directory contains generated contracts only; remove obsolete outputs too.
  find "$target" -name '*.go' -delete
  cp -R "$output"/. "$target"/
fi
