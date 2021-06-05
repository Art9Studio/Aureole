#!/bin/bash

set -e
set -u
set -x

docker pull aureolecloud/aureole-builder:latest || true
make build-docker-builder

docker pull aureolecloud/aureole:latest || true
make build-docker-image