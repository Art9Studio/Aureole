#!/bin/bash

set -e
set -u
set -x

docker build . -t twilio-stub:latest