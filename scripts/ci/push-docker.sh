#!/bin/bash

set -e
set -u
set -x

docker push aureolecloud/aureole-builder:latest

docker push aureolecloud/aureole:latest