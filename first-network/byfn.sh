#!/bin/bash

#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

# This script will orchestrate a sample end-to-end execution of the Hyperledger 
# Fabric network.
# (이 스크립트는 Hyperledger 페브릭 네트워크의 엔드 투 엔드 실행 샘플을 조정합니다.)
#
# The end-to-end verification provisions a sample Fabric network consisting of
# two organizations, each maintaining two peers, and a “solo” ordering service.
# (엔드 투 엔드 검증은 두 개의 조직을 유지 관리하는 두 개의 조직과 "솔로"주문 ​​서비스로 구성된 샘플 패브릭 네트워크를 제공합니다.)
#
# This verification makes use of two fundamental tools, which are necessary to
# create a functioning transactional network with digital signature validation
# and access control:
# (이 확인은 디지털 서명 유효성 검사 및 액세스 제어 기능이있는 트랜잭션 네트워크를 만드는 데 필요한 두 가지 기본 도구를 사용합니다.)
#
# * cryptogen - generates the x509 certificates used to identify and
#   authenticate the various components in the network.
#   (네트워크의 다양한 구성 요소를 식별하고 인증하는 데 사용되는 x509 인증서를 생성합니다.)
# * configtxgen - generates the requisite configuration artifacts for orderer
#   bootstrap and channel creation.
#   (주문자 부트 스트랩 및 채널 작성을 위해 필수 구성 아티팩트를 생성합니다.)
#
# Each tool consumes a configuration yaml file, within which we specify the topology
# of our network (cryptogen) and the location of our certificates for various
# configuration operations (configtxgen).  Once the tools have been successfully run,
# we are able to launch our network.  More detail on the tools and the structure of
# the network will be provided later in this document.  For now, let's get going...
# (각 도구는 네트워크의 토폴로지(cryptogen)와 다양한 구성 작업(configtxgen)에 대한 인증서의 위치를 지정하는 구성 yaml 파일을 사용합니다.
# (도구가 성공적으로 실행되면 네트워크를 시작할 수 있습니다. 도구 및 네트워크 구조에 대한 자세한 내용은이 문서 뒷부분에 나와 있습니다. 지금은 일단 진행하십시오.)

# prepending $PWD/../bin to PATH to ensure we are picking up the correct binaries
# this may be commented out to resolve installed version of tools if desired
# ($ PWD /../ bin을 PATH에 prepending하여 올바른 바이너리를 선택했는지 확인합니다. 원하는 경우 도구의 설치된 버전을 해결하기 위해 주석 처리 될 수 있습니다.)
export PATH=${PWD}/../bin:${PWD}:$PATH
export FABRIC_CFG_PATH=${PWD}

# Print the usage message(사용법 메시지를 출력합니다.)
function printHelp () {
  echo "Usage: "
  echo "  byfn.sh -m up|down|restart|generate [-c <channel name>] [-t <timeout>] [-d <delay>] [-f <docker-compose-file>] [-s <dbtype>]"
  echo "  byfn.sh -h|--help (print this message)"
  echo "    -m <mode> - one of 'up', 'down', 'restart' or 'generate'"
  echo "      - 'up' - bring up the network with docker-compose up"
  echo "      - 'down' - clear the network with docker-compose down"
  echo "      - 'restart' - restart the network"
  echo "      - 'generate' - generate required certificates and genesis block"
  echo "    -c <channel name> - channel name to use (defaults to \"mychannel\")"
  echo "    -t <timeout> - CLI timeout duration in microseconds (defaults to 10000)"
  echo "    -d <delay> - delay duration in seconds (defaults to 3)"
  echo "    -f <docker-compose-file> - specify which docker-compose file use (defaults to docker-compose-cli.yaml)"
  echo "    -s <dbtype> - the database backend to use: goleveldb (default) or couchdb"
  echo
  echo "Typically, one would first generate the required certificates and "
  echo "genesis block, then bring up the network. e.g.:"
  echo
  echo "	byfn.sh -m generate -c mychannel"
  echo "	byfn.sh -m up -c mychannel -s couchdb"
  echo "	byfn.sh -m down -c mychannel"
  echo
  echo "Taking all defaults:"
  echo "	byfn.sh -m generate"
  echo "	byfn.sh -m up"
  echo "	byfn.sh -m down"
}

# Ask user for confirmation to proceed(사용자에게 계속할 것인지 묻습니다.)
function askProceed () {
  read -p "Continue (y/n)? " ans
  case "$ans" in
    y|Y )
      echo "proceeding ..."
    ;;
    n|N )
      echo "exiting..."
      exit 1
    ;;
    * )
      echo "invalid response"
      askProceed
    ;;
  esac
}

# Obtain CONTAINER_IDS and remove them(CONTAINER_IDS 얻기 및 제거)
# TODO Might want to make this optional - could clear other containers
# (TODO 항목을 선택적으로 설정하려는 경우 - 다른 컨테이너를 지울 수 있습니다.)
function clearContainers () {
  CONTAINER_IDS=$(docker ps -aq)
  if [ -z "$CONTAINER_IDS" -o "$CONTAINER_IDS" == " " ]; then
    echo "---- No containers available for deletion ----"
  else
    docker rm -f $CONTAINER_IDS
  fi
}

# Delete any images that were generated as a part of this setup(이 설정의 일부로 생성된 모든 이미지 삭제)
# specifically the following images are often left behind:(특히 다음과 같은 이미지가 종종 남습니다.)
# TODO list generated image naming patterns(목록에 생성된 이미지 이름 지정 패턴)
function removeUnwantedImages() {
  DOCKER_IMAGE_IDS=$(docker images | grep "dev\|none\|test-vp\|peer[0-9]-" | awk '{print $3}')
  if [ -z "$DOCKER_IMAGE_IDS" -o "$DOCKER_IMAGE_IDS" == " " ]; then
    echo "---- No images available for deletion ----"
  else
    docker rmi -f $DOCKER_IMAGE_IDS
  fi
}

# Generate the needed certificates, the genesis block and start the network.
# (필요한 인증서와 제네시스 블록을 생성하고 네트워크를 시작하십시오.)
function networkUp () {
  # generate artifacts if they don't exist(인증서와 제네시스 블록이 존재하지 않으면 아티팩트를 생성한다.)
  if [ ! -d "crypto-config" ]; then
    generateCerts
    replacePrivateKey
    generateChannelArtifacts
  fi
  if [ "${IF_COUCHDB}" == "couchdb" ]; then
      CHANNEL_NAME=$CHANNEL_NAME TIMEOUT=$CLI_TIMEOUT DELAY=$CLI_DELAY docker-compose -f $COMPOSE_FILE -f $COMPOSE_FILE_COUCH up -d 2>&1
  else
      CHANNEL_NAME=$CHANNEL_NAME TIMEOUT=$CLI_TIMEOUT DELAY=$CLI_DELAY docker-compose -f $COMPOSE_FILE up -d 2>&1
  fi
  if [ $? -ne 0 ]; then
    echo "ERROR !!!! Unable to start network"
    docker logs -f cli
    exit 1
  fi
  docker logs -f cli
}

# Tear down running network(실행중인 네트워크 종료)
function networkDown () {
  docker-compose -f $COMPOSE_FILE down
  docker-compose -f $COMPOSE_FILE -f $COMPOSE_FILE_COUCH down
  # Don't remove containers, images, etc if restarting(다시 시작한다면, 컨테이너 이미지 등을 제거하지 마십시오.)
  if [ "$MODE" != "restart" ]; then
    #Cleanup the chaincode containers(체인 코드 컨테이너 정리)
    clearContainers
    #Cleanup images(이미지 삭제)
    removeUnwantedImages
    # remove orderer block and other channel configuration transactions and certs(주문자 블록과 기타 채널 구성 트랜잭션과 인증서 제거)
    rm -rf channel-artifacts/*.block channel-artifacts/*.tx crypto-config
    # remove the docker-compose yaml file that was customized to the example(예제에 맞게 사용자 정의된 docker-compose yaml 파일을 제거하십시오.)
    rm -f docker-compose-e2e.yaml
  fi
}

# Using docker-compose-e2e-template.yaml, replace constants with private key file names
# generated by the cryptogen tool and output a docker-compose.yaml specific to this
# configuration
# (docker-compose-e2e-template.yaml을 사용하여 cryptogen 도구로 생성 된 프라이빗 키 파일 이름으로
#  상수를 대체하고 이 구성과 관련된 docker-compose.yaml을 출력하십시오.)
function replacePrivateKey () {
  # sed on MacOSX does not support -i flag with a null extension. We will use
  # 't' for our back-up's extension and depete it at the end of the function
  # (MacOSX에서 sed는 널 확장자를 가진 -i 플래그를 지원하지 않습니다.
  # 우리는 백업의 확장을 위해 't'를 사용할 것이고 함수의 끝에 그것을 빼낼 것입니다.)
  ARCH=`uname -s | grep Darwin`
  if [ "$ARCH" == "Darwin" ]; then
    OPTS="-it"
  else
    OPTS="-i"
  fi

  # Copy the template to the file that will be modified to add the private key
  # (프라이빗 키를 추가하기 위해 수정할 파일에 템플릿을 복사하십시오.)
  cp docker-compose-e2e-template.yaml docker-compose-e2e.yaml

  # The next steps will replace the template's contents with the
  # actual values of the private key file names for the two CAs.
  # (다음 단계에서는 템플릿의 내용을 두 CA의 프라이빗 키 파일 이름의 실제 값으로 바꿉니다.)
  CURRENT_DIR=$PWD
  cd crypto-config/peerOrganizations/org1.example.com/ca/
  PRIV_KEY=$(ls *_sk)
  cd $CURRENT_DIR
  sed $OPTS "s/CA1_PRIVATE_KEY/${PRIV_KEY}/g" docker-compose-e2e.yaml
  cd crypto-config/peerOrganizations/org2.example.com/ca/
  PRIV_KEY=$(ls *_sk)
  cd $CURRENT_DIR
  sed $OPTS "s/CA2_PRIVATE_KEY/${PRIV_KEY}/g" docker-compose-e2e.yaml
  # If MacOSX, remove the temporary backup of the docker-compose file
  if [ "$ARCH" == "Darwin" ]; then
    rm docker-compose-e2e.yamlt
  fi
}

# We will use the cryptogen tool to generate the cryptographic material (x509 certs)
# for our various network entities.  The certificates are based on a standard PKI
# implementation where validation is achieved by reaching a common trust anchor.
# (cryptogen 도구를 사용하여 다양한 네트워크 엔터티에 대한 암호 자료 (x509 certs)를 생성합니다.
#  인증서는 공통 PKI 구현을 기반으로하며 여기에서는 공통 트러스트 앵커에 도달하여 유효성을 검사합니다.)
#
# Cryptogen consumes a file - ``crypto-config.yaml`` - that contains the network
# topology and allows us to generate a library of certificates for both the
# Organizations and the components that belong to those Organizations.  Each
# Organization is provisioned a unique root certificate (``ca-cert``), that binds
# specific components (peers and orderers) to that Org.  Transactions and communications
# within Fabric are signed by an entity's private key (``keystore``), and then verified
# by means of a public key (``signcerts``).  You will notice a "count" variable within
# this file.  We use this to specify the number of peers per Organization; in our
# case it's two peers per Org.  The rest of this template is extremely
# self-explanatory.
# (Cryptogen은 네트워크 토폴로지를 포함하고있는``crypto-config.yaml`` 파일을 사용하며 조직과
#  그 조직에 속한 구성 요소 모두에 대한 인증서 라이브러리를 생성 할 수 있습니다. 각 조직은 특정 구성 요소(피어 및 ​​주문자)를
# 해당 조직에 바인딩하는 고유 한 루트 인증서 ( "ca-cert")를 제공합니다.
# Fabric 내의 트랜잭션과 통신은 엔티티의 프라이빗 키 ( "keystore")에 의해 서명 된 다음 공개 키 ( "signcerts")를 통해 검증됩니다.
# 이 파일에는 "count"변수가 있습니다. 우리는이를 사용하여 조직 당 피어의 수를 지정합니다.
# 우리의 경우 Org 당 두 명의 동료가 있습니다. 이 템플릿의 나머지 부분은 매우 자명합니다.)
#
# After we run the tool, the certs will be parked in a folder titled ``crypto-config``.
# (이 도구를 실행하면 certs는 "crypto-config"라는 폴더에 보관됩니다.)

# Generates Org certs using cryptogen tool(cryptogen 도구를 사용하여 조직 certs를 생성합니다.)
function generateCerts (){
  which cryptogen
  if [ "$?" -ne 0 ]; then
    echo "cryptogen tool not found. exiting"
    exit 1
  fi
  echo
  echo "##########################################################"
  echo "##### Generate certificates using cryptogen tool #########"
  echo "##########################################################"

  cryptogen generate --config=./crypto-config.yaml
  if [ "$?" -ne 0 ]; then
    echo "Failed to generate certificates..."
    exit 1
  fi
  echo
}

# The `configtxgen tool is used to create four artifacts: orderer **bootstrap
# block**, fabric **channel configuration transaction**, and two **anchor
# peer transactions** - one for each Peer Org.
# (configtxgen 도구는 주문자 ** 부트 스트랩 블록 **, 패브릭 ** 채널 구성 트랜잭션 **,
#  두 개의 앵커 피어 트랜잭션 ** 두 개의 아티팩트를 생성하는 데 사용됩니다. ** - 각각의 피어 조직을 위한 것입니다.)
#
# The orderer block is the genesis block for the ordering service, and the
# channel transaction file is broadcast to the orderer at channel creation
# time.  The anchor peer transactions, as the name might suggest, specify each
# Org's anchor peer on this channel.
# (발주자 블록은 주문 서비스를 위한 제네시스 블록이며, 채널 트랜잭션 파일은 채널 생성시 발주자에게 방송됩니다.
#  앵커 피어 트랜잭션은 이름에서 알 수 있듯이 이 채널에서 각 조직의 앵커 피어를 지정합니다.)
#
# Configtxgen consumes a file - ``configtx.yaml`` - that contains the definitions
# for the sample network. There are three members - one Orderer Org (``OrdererOrg``)
# and two Peer Orgs (``Org1`` & ``Org2``) each managing and maintaining two peer nodes.
# This file also specifies a consortium - ``SampleConsortium`` - consisting of our
# two Peer Orgs.  Pay specific attention to the "Profiles" section at the top of
# this file.  You will notice that we have two unique headers. One for the orderer genesis
# block - ``TwoOrgsOrdererGenesis`` - and one for our channel - ``TwoOrgsChannel``.
# These headers are important, as we will pass them in as arguments when we create
# our artifacts.  This file also contains two additional specifications that are worth
# noting.  Firstly, we specify the anchor peers for each Peer Org
# (``peer0.org1.example.com`` & ``peer0.org2.example.com``).  Secondly, we point to
# the location of the MSP directory for each member, in turn allowing us to store the
# root certificates for each Org in the orderer genesis block.  This is a critical
# concept. Now any network entity communicating with the ordering service can have
# its digital signature verified.
# (Configtxgen은 샘플 네트워크의 정의를 포함하는 파일``configtx.yaml``을 사용합니다.
# Orderer Org (Orderer Org)과 Orgerer Org (Orgerer Org)는 세 개의 구성원으로 구성되어 있으며
# 각각 두 개의 피어 노드를 관리하고 유지 관리하는 두 개의 Peer Orgs (Org1 및 Org2)가 있습니다.
# 이 파일은 또한 두 개의 Peer Orgs로 구성된 컨소시엄 인 "SampleConsortium"을 지정합니다.
# 이 파일 맨 위에있는 "프로필"섹션에 특히주의하십시오. 두 개의 고유한 헤더가 있음을 알 수 있습니다.
# 발주자 생성 블록 "TwoOrgsOrdererGenesis"과 채널 "TwoOrgsChannel"중 하나입니다.)
#
# This function will generate the crypto material and our four configuration
# artifacts, and subsequently output these files into the ``channel-artifacts``
# folder.
# (이 함수는 암호 자료와 네 가지 구성 아티팩트를 생성 한 후 이 파일을 ``channel-artifacts``폴더로 출력합니다.)
#
# If you receive the following warning, it can be safely ignored:
# (다음과 같은 경고 메시지가 나타나면 무시해도됩니다.)
#
# [bccsp] GetDefault -> WARN 001 Before using BCCSP, please call InitFactories(). Falling back to bootBCCSP.
#
# You can ignore the logs regarding intermediate certs, we are not using them in
# this crypto implementation.
# (중간 인증서에 관한 로그는 무시할 수 있으며,이 암호화 구현에서는 사용하지 않습니다.)

# Generate orderer genesis block, channel configuration transaction and
# anchor peer update transactions
# (발주자 생성 블록, 채널 구성 트랜잭션과 앵커 피어 업데이트 트랜잭션을 생성합니다.)
function generateChannelArtifacts() {
  which configtxgen
  if [ "$?" -ne 0 ]; then
    echo "configtxgen tool not found. exiting"
    exit 1
  fi

  echo "##########################################################"
  echo "#########  Generating Orderer Genesis block ##############"
  echo "##########################################################"
  # Note: For some unknown reason (at least for now) the block file can't be
  # named orderer.genesis.block or the orderer will fail to launch!
  # (알 수없는 이유로(적어도, 지금은) 블록 파일의 이름을 orderer.genesis.block으로 지정할 수 없거나 주문자가 실행되지 않습니다!)
  configtxgen -profile TwoOrgsOrdererGenesis -outputBlock ./channel-artifacts/genesis.block
  if [ "$?" -ne 0 ]; then
    echo "Failed to generate orderer genesis block..."
    exit 1
  fi
  echo
  echo "#################################################################"
  echo "### Generating channel configuration transaction 'channel.tx' ###"
  echo "#################################################################"
  configtxgen -profile TwoOrgsChannel -outputCreateChannelTx ./channel-artifacts/channel.tx -channelID $CHANNEL_NAME
  if [ "$?" -ne 0 ]; then
    echo "Failed to generate channel configuration transaction..."
    exit 1
  fi

  echo
  echo "#################################################################"
  echo "#######    Generating anchor peer update for Org1MSP   ##########"
  echo "#################################################################"
  configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org1MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org1MSP
  if [ "$?" -ne 0 ]; then
    echo "Failed to generate anchor peer update for Org1MSP..."
    exit 1
  fi

  echo
  echo "#################################################################"
  echo "#######    Generating anchor peer update for Org2MSP   ##########"
  echo "#################################################################"
  configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate \
  ./channel-artifacts/Org2MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org2MSP
  if [ "$?" -ne 0 ]; then
    echo "Failed to generate anchor peer update for Org2MSP..."
    exit 1
  fi
  echo
}

# Obtain the OS and Architecture string that will be used to select the correct
# native binaries for your platform
# (해당 플랫폼에 맞는 올바른 원시 바이너리를 선택하는 데 사용할 OS 및 아키텍처 문자열을 얻습니다.)
OS_ARCH=$(echo "$(uname -s|tr '[:upper:]' '[:lower:]'|sed 's/mingw64_nt.*/windows/')-$(uname -m | sed 's/x86_64/amd64/g')" | awk '{print tolower($0)}')
# timeout duration - the duration the CLI should wait for a response from
# another container before giving up
# (timeout duration - CLI가 다른 컨테이너의 응답을 기다려야 만하는 시간입니다.)
CLI_TIMEOUT=10000
#default for delay(지연을 위한 기본값)
CLI_DELAY=3
# channel name defaults to "mychannel"(채널 이름의 기본값은 "mychannel"입니다.)
CHANNEL_NAME="mychannel"
# use this as the default docker-compose yaml definition
# (이것을 기본 docker-compose yaml 정의로 사용하십시오.)
COMPOSE_FILE=docker-compose-cli.yaml
#
COMPOSE_FILE_COUCH=docker-compose-couch.yaml

# Parse commandline args(구문 커맨드라인 인수)
while getopts "h?m:c:t:d:f:s:" opt; do
  case "$opt" in
    h|\?)
      printHelp
      exit 0
    ;;
    m)  MODE=$OPTARG
    ;;
    c)  CHANNEL_NAME=$OPTARG
    ;;
    t)  CLI_TIMEOUT=$OPTARG
    ;;
    d)  CLI_DELAY=$OPTARG
    ;;
    f)  COMPOSE_FILE=$OPTARG
    ;;
    s)  IF_COUCHDB=$OPTARG
    ;;
  esac
done

# Determine whether starting, stopping, restarting or generating for announce
# (공지를 위해 시작, 중지, 재시작 또는 생성 여부를 결정하십시오.)
if [ "$MODE" == "up" ]; then
  EXPMODE="Starting"
  elif [ "$MODE" == "down" ]; then
  EXPMODE="Stopping"
  elif [ "$MODE" == "restart" ]; then
  EXPMODE="Restarting"
  elif [ "$MODE" == "generate" ]; then
  EXPMODE="Generating certs and genesis block for"
else
  printHelp
  exit 1
fi

# Announce what was requested(요청 된 것을 공지하십시오.)

  if [ "${IF_COUCHDB}" == "couchdb" ]; then
        echo
        echo "${EXPMODE} with channel '${CHANNEL_NAME}' and CLI timeout of '${CLI_TIMEOUT}' using database '${IF_COUCHDB}'"
  else
        echo "${EXPMODE} with channel '${CHANNEL_NAME}' and CLI timeout of '${CLI_TIMEOUT}'"
  fi
# ask for confirmation to proceed(계속할 것인지 물어 봅니다.)
askProceed

#Create the network using docker compose(docker 작성을 사용하여 네트워크를 만듭니다.)
if [ "${MODE}" == "up" ]; then
  networkUp
  elif [ "${MODE}" == "down" ]; then ## Clear the network(네트워크 지우기)
  networkDown
  elif [ "${MODE}" == "generate" ]; then ## Generate Artifacts(이슈 생성)
  generateCerts
  replacePrivateKey
  generateChannelArtifacts
  elif [ "${MODE}" == "restart" ]; then ## Restart the network(네트워크 다시 시작)
  networkDown
  networkUp
else
  printHelp
  exit 1
fi
