#!/bin/bash
set -e

ROOT_MODULE="github.com/k1v4/drip_mate"
OUTPUT_FILE="mockery.yml"

# Шапка YAML
cat > "$OUTPUT_FILE" <<EOF
quiet: false
log-level: error
disable-version-string: true
with-expecter: true
disable-func-mocks: true
outpkg: mocks

packages:
EOF

# Находим все папки internal и получаем уникальные сервисы
find . -type d -name "internal" | while read -r internal_dir; do
    service=$(basename "$(dirname "$internal_dir")")

    # Игнорируем корневой ./internal, если такой есть
    if [ "$service" = "." ] || [ "$service" = "" ]; then
        continue
    fi

    go_path="$ROOT_MODULE/$service/internal"
    dir_path="$service/internal/mocks/{{replaceAll (replaceAll (replace .InterfaceDirRelative \"$service/internal/\" \"\" -1) \"internal\" \"internal_\") \"vendor/\" \"\"}}"

    cat >> "$OUTPUT_FILE" <<EOF
  $go_path:
    config:
      all: true
      recursive: true
      exclude:
        - "mocks"
        - "pkg/tests"
      dir: $dir_path
      outpkg: mocks
      mockname: "{{ .InterfaceName | camelcase }}"
EOF

done

echo "mockery.yml generated successfully."
