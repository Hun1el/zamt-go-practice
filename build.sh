#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== Очистка старых скомпилированных файлов ==="

for gofile in "$SCRIPT_DIR"/*.go; do
    if [ -f "$gofile" ]; then
        basename=$(basename "$gofile" .go)
        progname="${basename%%_*}"
        
        if [ -f "$SCRIPT_DIR/$progname" ] && [ ! -f "$SCRIPT_DIR/$progname.go" ]; then
            rm -f "$SCRIPT_DIR/$progname"
            echo "Удалён старый $SCRIPT_DIR/$progname"
        fi
    fi
done

export GO111MODULE=off

echo ""
echo "=== Установка прав на исходники ==="
chmod -R 777 "$SCRIPT_DIR" 2>/dev/null

echo ""
echo "=== Компиляция ==="
SUCCESS=0
FAILED=0

declare -A programs

for gofile in "$SCRIPT_DIR"/*.go; do
    if [ -f "$gofile" ]; then
        basename=$(basename "$gofile" .go)
        progname="${basename%%_*}"
        
        if [ -z "${programs[$progname]}" ]; then
            programs[$progname]=1
        fi
    fi
done

for progname in "${!programs[@]}"; do
    echo -n "Компиляция $progname... "
    
    FILES=""
    if [ -f "$SCRIPT_DIR/${progname}.go" ]; then
        FILES="$SCRIPT_DIR/${progname}.go"
    fi
    
    for helper in "$SCRIPT_DIR"/${progname}_*.go; do
        if [ -f "$helper" ]; then
            FILES="$FILES $helper"
        fi
    done
    
    if [ -z "$FILES" ]; then
        echo "ОШИБКА: файлы не найдены"
        ((FAILED++))
        continue
    fi
    
    if (cd "$SCRIPT_DIR" && go build -o "$progname" $FILES 2>&1); then
        if [ -f "$SCRIPT_DIR/$progname" ]; then
            chmod 777 "$SCRIPT_DIR/$progname"
            echo "OK"
            ((SUCCESS++))
        else
            echo "ОШИБКА: файл не создан"
            ((FAILED++))
        fi
    else
        echo "ОШИБКА компиляции"
        ((FAILED++))
    fi
done

echo ""
echo "=== Установка прав на все исполняемые файлы ==="

for progname in "${!programs[@]}"; do
    if [ -f "$SCRIPT_DIR/$progname" ]; then
        chmod 777 "$SCRIPT_DIR/$progname"
    fi
done

echo ""
echo "============================================"
echo "Готово!"
echo "Успешно: $SUCCESS"
echo "Ошибок: $FAILED"
echo "============================================"
echo ""
echo "Скомпилированные файлы:"

for progname in "${!programs[@]}"; do
    if [ -f "$SCRIPT_DIR/$progname" ]; then
        ls -lh "$SCRIPT_DIR/$progname"
    fi
done
