#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#
# Exit on first error, print all commands.
set -ev

# Shut down the Docker containers that might be currently running.
# (현재 실행중인 Docker 컨테이너를 종료하십시오.)
docker-compose -f docker-compose.yml stop
