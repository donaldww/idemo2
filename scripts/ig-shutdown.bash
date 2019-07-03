#!/usr/bin/env bash

echo "Shutting down idemo"
pkill cockroach >/dev/null
sleep 1
pkill cockroach >/dev/null
sleep 1
# ps -axf | grep "[c]ockroach"

pgrep -f "cockroach"
