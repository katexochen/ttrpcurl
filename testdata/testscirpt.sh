#!/usr/bin/env bash

set -euo pipefail

function cleanup() {
    kill -- $bgpid
    rm -f  ./ttrpcurl
    rm -f /tmp/ttrpc-test.sock
}

trap cleanup EXIT

if ! command -v coco-ttrpc-cli &> /dev/null; then
    go install github.com/katexochen/coco-ttrpc-cli@latest
fi

coco-ttrpc-cli --socket /tmp/ttrpc-test.sock attestation-agent server &
bgpid=$!

sleep 1

go build ../cmd/ttrpcurl

data=$(cat data.json)

./ttrpcurl \
    --proto getresource.proto \
    --data "${data}" \
    /tmp/ttrpc-test.sock \
    getresource.GetResourceService.GetResource

echo "${data}" \
| ./ttrpcurl \
    --proto getresource.proto \
    --plaintext \
    --data @ \
    /tmp/ttrpc-test.sock \
    getresource.GetResourceService.GetResource
