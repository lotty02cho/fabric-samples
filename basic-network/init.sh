#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

# Exit on first error, print all commands.(첫 번째 오류가 발생하면 종료하고 모든 명령을 인쇄하십시오.)
set -ev
# delete previous creds(이전 creds 삭제)
rm -rf ~/.hfc-key-store/*

# copy peer admin credentials into the keyValStore(피어 관리자 자격 증명을 keyValStore에 복사)
mkdir -p ~/.hfc-key-store
cp creds/* ~/.hfc-key-store
