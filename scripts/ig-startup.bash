#!/usr/bin/env bash

# Start cockroachdb

cd "$IGHOME/data" || exit

echo ; echo "Starting Cockroachdb"; echo

cockroach start  --certs-dir="$IGHOME/certs" --listen-addr=localhost > /dev/null 2>&1 &

cockroach start   --certs-dir="$IGHOME/certs" --store=node2 --listen-addr=localhost:26258 \
--http-addr=localhost:8081 --join=localhost:26257  > /dev/null 2>&1 &

cockroach start  --certs-dir="$IGHOME/certs" --store=node3 --listen-addr=localhost:26259 \
--http-addr=localhost:8082 --join=localhost:26257 > /dev/null 2>&1  &

sleep 1

# Show cockroach nodes
# ps -axf | grep "[c]ockroach | grep -v grep"

pgrep -f "cockroach"