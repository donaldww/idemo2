#!/usr/bin/env bash

IPATH=~/Library/Infinigon

echo "Starting idemo"

# Clear the old IPATH directory.
if [[ -d ${IPATH} ]]
then
	echo "Removing old ${IPATH}"
	yes | rm -r ${IPATH}
fi

# Create new IPATH DIRECTORY Structure
echo "Creating new ${IPATH}"
mkdir -p ${IPATH}/bin
mkdir -p ${IPATH}/db
mkdir -p ${IPATH}/tmp

# Start cockroachdb
cd ${IPATH}/db
echo ; echo "Starting cockroach db"
cockroach start --insecure --listen-addr=localhost 2>1 > /dev/null &
cockroach start --insecure --store=node2 --listen-addr=localhost:26258 --http-addr=localhost:8081 \
--join=localhost:26257 2>1 > /dev/null &
cockroach start --insecure --store=node3 --listen-addr=localhost:26259 --http-addr=localhost:8082 \
--join=localhost:26257 2>1 > /dev/null &

# Show cockroach nodes
ps -axf | grep "[c]ockroach"

# TODO: setup bin directory

cd ${IPATH}/tmp
echo ; echo "Ready"
