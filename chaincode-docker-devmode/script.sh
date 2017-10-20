#!/bin/bash
# Copyright London Stock Exchange Group All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
set -e
# This script expedites the chaincode development process by automating the
# requisite channel create/join commands
# (이 스크립트는 필수 채널 create/join 명령을 자동화하여 체인 코드 개발 프로세스를 신속하게 처리합니다.)

# We use a pre-generated orderer.block and channel transaction artifact (myc.tx),
# both of which are created using the configtxgen tool
# (우리는 미리 생성된 orderer.block 및 채널 트랜잭션 아티팩트 (myc.tx)를 사용하며, 둘 다 configtxgen 도구를 사용하여 생성됩니다.)

# first we create the channel against the specified configuration in myc.tx
# this call returns a channel configuration block - myc.block - to the CLI container
# (먼저 myc.tx의 지정된 구성에 대해 채널을 만듭니다.
# 이 호출은 채널 구성 블록 인 myc.block을 CLI 컨테이너에 반환합니다.)
peer channel create -c myc -f myc.tx -o orderer:7050

# now we will join the channel and start the chain with myc.block serving as the
# channel's first block (i.e. the genesis block)
# 이제 우리는 채널에 가입하고 채널의 첫 번째 블록 (즉, genesis 블록) 역할을 하는 myc.block을 사용하여 체인을 시작합니다.
peer channel join -b myc.block

# Now the user can proceed to build and start chaincode in one terminal
# And leverage the CLI container to issue install instantiate invoke query commands in another
# (이제 사용자는 하나의 터미널에서 체인 코드를 작성하고 시작할 수 있습니다.
# 그리고 CLI 컨테이너를 활용하여 다른 인스턴스에서 install query execute 명령을 설치합니다.)

#we should have bailed if above commands failed.
#we are here, so they worked
#(위의 명령이 실패했다면, 이 스크립트를 빠져나와야합니다.
#그대로 있다면, 이것들은 작동됐습니다.)
sleep 600000
exit 0
