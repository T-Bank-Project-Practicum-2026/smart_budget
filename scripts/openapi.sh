#!/bin/sh
set -eu

mode=${1:-generate}
case "$mode" in
  generate|check) ;;
  *) echo "Использование: $0 [generate|check]" >&2; exit 2 ;;
esac
output=$(mktemp -d)
trap 'rm -rf "$output"' EXIT HUP INT TERM

swag init \
  --v3.1 --generalInfo doc.go \
  --dir internal/server,internal/auth,internal/push,internal/runtime \
  --parseInternal --requiredByDefault --outputTypes yaml \
  --output "$output"
{ echo '# Сгенерировано командой. Не редактируйте вручную.'; cat "$output/swagger.yaml"; } > "$output/openapi.yaml"

if [ "$mode" = check ]; then
  diff -u ../api/openapi.yaml "$output/openapi.yaml" || {
    echo 'OpenAPI устарел. Запустите make swagger.' >&2
    exit 1
  }
else
  cp "$output/openapi.yaml" ../api/openapi.yaml
fi
