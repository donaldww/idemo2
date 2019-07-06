#!/usr/local/bin/zsh

echo "Shutting down cockroachdb"
pkill cockroach >/dev/null
sleep 1
pkill cockroach >/dev/null
sleep 1

pgrep -fl cockroach
