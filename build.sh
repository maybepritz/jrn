#!/bin/bash

# Название выходного бинарника
APP_NAME="jrn"
OUTPUT_DIR="./bin"
ENTRY_POINT="cmd/jrn/main.go"

# Очищаем старую папку bin
rm -rf $OUTPUT_DIR
mkdir -p $OUTPUT_DIR

# Список платформ: OS/ARCH
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
    "windows/arm64"
)

echo "🚀 Начало сборки $APP_NAME..."

for PLATFORM in "${PLATFORMS[@]}"
do
    OS="${PLATFORM%/*}"
    ARCH="${PLATFORM#*/}"
    OUT="${OUTPUT_DIR}/${APP_NAME}-${OS}-${ARCH}"
    
    if [ "$OS" == "windows" ]; then
        OUT="${OUT}.exe"
    fi

    echo "📦 Сборка под ${OS}/${ARCH} -> ${OUT}"
    env GOOS=$OS GOARCH=$ARCH CGO_ENABLED=0 go build -ldflags="-s -w" -o $OUT $ENTRY_POINT
    if [ $? -ne 0 ]; then
        echo "❌ Ошибка при сборке под ${OS}/${ARCH}"
        exit 1
    fi
done

echo "✅ Все сборки успешно созданы в папке ${OUTPUT_DIR}/"
ls -lh $OUTPUT_DIR