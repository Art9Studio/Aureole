#!/bin/bash

set -e
set -u
set -x

docker build --no-cache . -t aureole:latest