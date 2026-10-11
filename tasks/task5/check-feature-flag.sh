#!/bin/bash

set -e

URL="${URL:-http://localhost:9090}"

echo "▶️ Проверка Feature Flag (X-Feature-Enabled: true)..."

# Отправляем запрос с заголовком, чтобы маршрутизировать трафик на `v2`
echo "With X-Feature-Enabled: true"
for i in 1 2 3 4 5; do
    echo "  /version -> $(curl -s -H "X-Feature-Enabled: true" "$URL/version")"
done
echo "  /feature -> $(curl -s -H "X-Feature-Enabled: true" "$URL/feature")"

echo "Without the header (canary: mostly v1, where /feature does not exist)"
echo "  /feature -> HTTP $(curl -s -o /dev/null -w '%{http_code}' "$URL/feature")"

echo "Client cannot pick v2 by the internal header (the EnvoyFilter drops it)"
for i in $(seq 1 20); do curl -s -H "x-booking-version: v2" "$URL/version"; echo; done | sort | uniq -c
