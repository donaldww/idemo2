#!/usr/bin/env bash

pkill cockroach
echo "Shutting down idemo"
sleep 1
ps -axf | grep "[c]ockroach"
