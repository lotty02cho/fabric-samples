#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#
# Exit on first error, print all commands.
set -e

# Shut down the Docker containers for the system tests.
# (시스템 테스트를 위해 Docker 컨테이너를 종료하십시오.)
docker-compose -f docker-compose.yml kill && docker-compose -f docker-compose.yml down

# remove the local state(로컬 상태를 제거합니다.)
rm -f ~/.hfc-key-store/*

# remove chaincode docker images(체인 코드 고정 이미지 제거합니다.)
docker rmi $(docker images dev-* -q)

# Your system is now clean(귀하의 시스템은 이제 깨끗합니다.)
