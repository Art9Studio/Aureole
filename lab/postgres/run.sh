#!/bin/bash

set -e
set -u
set -x

DIR_PATH=$(dirname "$0")

docker-compose -f "$DIR_PATH"/docker-compose.yml up --remove-orphans -d