#!/usr/bin/env bash

# Start cockroachdb
cd $(IDATA)
echo ; echo "Starting cockroachdb"; echo
cockroach start --insecure --listen-addr=localhost 2>1 > /dev/null &
cockroach start --insecure --store=node2 --listen-addr=localhost:26258 --http-addr=localhost:8081 \
--join=localhost:26257 2>1 > /dev/null &
cockroach start --insecure --store=node3 --listen-addr=localhost:26259 --http-addr=localhost:8082 \
--join=localhost:26257 2>1 > /dev/null &

# Show cockroach nodes
ps -axf | grep "[c]ockroach | grep -v grep"
