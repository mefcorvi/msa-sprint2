#!/bin/bash

set -e

URL="${URL:-http://localhost:9090}"
REQUESTS="${REQUESTS:-20}"

echo "▶️ Testing fallback route..."

results=$(for i in $(seq 1 "$REQUESTS"); do
    curl -s -w ' HTTP %{http_code}' "$URL/version"
    echo
done)
echo "$results" | sort | uniq -c

if echo "$results" | grep -qv 'HTTP 200'; then
    echo "Fallback failed: some requests were not served"
    exit 1
fi
echo "Fallback route working"
