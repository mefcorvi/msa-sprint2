#!/bin/bash

set -e

URL="${URL:-http://localhost:9090}"
REQUESTS="${REQUESTS:-100}"

echo "▶️ Checking canary release (90% v1, 10% v2)..."

# Посылаем 100 запросов; /version отвечает версией экземпляра, который обработал запрос
for i in $(seq 1 "$REQUESTS")
do
    curl -s "$URL/version"
    echo
done | sort | uniq -c
