#!/usr/bin/env bash

pkill cockroach > /dev/null
echo "Shutting down idemo"
sleep 1
ps -axf | grep "[c]ockroach"
