#!/usr/bin/env bash

IPATH=~/Library/Infinigon

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

