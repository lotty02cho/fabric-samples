개발 모드 사용(Using dev mode)
==============

일반적으로 체인 코드는 피어에 의해 시작되고 유지 관리됩니다. 그러나 "dev 모드"에서는 사용자가 체인 코드를 작성하고 시작합니다.이 모드는 rapid code / build / run / debug 사이클 전환을 위해 체인 코드 개발 단계에서 유용합니다.
우리는 샘플 개발 네트워크를 위해 사전 생성된 orderer와 채널 산출물을 활용하여 "dev mode"를 시작합니다. 따라서 사용자는 즉시 체인 코드 컴파일 및 호출 호출 프로세스로 이동할 수 있습니다.


패브릭 샘플 설치(Install Fabric Samples) 
----------------------

이미 그렇게하지 않았다면 : doc :`samples`을 설치하십시오.

fabric-samples 복제본의 chaincode-docker-devmode 디렉토리로 이동합니다.

.. code:: bash

  cd chaincode-docker-devmode

도커 이미지 다운로드(Download docker images) 
^^^^^^^^^^^^^^^^^^^^^^

제공된 도커 작성 스크립트에 대해 "dev 모드"를 실행하려면 4 개의 도커 이미지가 필요합니다. fabric-samples repo 복제본 을 설치하고 : ref :`download-platform-specific-binaries` 지시 사항을 따르면, 필요한 Docker 이미지가 로컬에 설치되어 있어야합니다.

.. 노트:: 이미지를 수동으로 가져 오기로 선택한 경우, 다시 ``latest`` 태그를 지정해야합니다.

실행하여 ``docker images`` 해당 지역의 부두 노동자 레지스트리를 공개하는 명령을 사용합니다. 다음과 비슷한 내용이 표시됩니다.

.. code:: bash

  docker images
  REPOSITORY                     TAG                                  IMAGE ID            CREATED             SIZE
  hyperledger/fabric-tools       latest                               e09f38f8928d        4 hours ago         1.32 GB
  hyperledger/fabric-tools       x86_64-1.0.0-rc1-snapshot-f20846c6   e09f38f8928d        4 hours ago         1.32 GB
  hyperledger/fabric-orderer     latest                               0df93ba35a25        4 hours ago         179 MB
  hyperledger/fabric-orderer     x86_64-1.0.0-rc1-snapshot-f20846c6   0df93ba35a25        4 hours ago         179 MB
  hyperledger/fabric-peer        latest                               533aec3f5a01        4 hours ago         182 MB
  hyperledger/fabric-peer        x86_64-1.0.0-rc1-snapshot-f20846c6   533aec3f5a01        4 hours ago         182 MB
  hyperledger/fabric-ccenv       latest                               4b70698a71d3        4 hours ago         1.29 GB
  hyperledger/fabric-ccenv       x86_64-1.0.0-rc1-snapshot-f20846c6   4b70698a71d3        4 hours ago         1.29 GB

.. 노트:: ref :`download-platform-specific-binaries`를 통해 이미지를 검색하면 추가 이미지가 나열됩니다. 그러나 우리는이 네 가지에만 관심이 있습니다.

N이제 세 개의 터미널을 열고 ``chaincode-docker-devmode`` 각각 의 디렉토리로 이동하십시오 .

터미널 1 - 네트워크 시작(Terminal 1 - Start the network) 
------------------------------

.. code:: bash

    docker-compose -f docker-compose-simple.yaml up

위의 예는 SingleSampleMSPSolo 주문자 프로필로 네트워크를 시작하고 "dev mode"에서 피어를 시작합니다. 또한 체인 코드 환경을위한 컨테이너와 체인 코드와 상호 작용하는 두 개의 CLI 컨테이너를 추가로 생성합니다. create and join 채널에 대한 명령은 CLI 컨테이너에 내장되어 있으므로 즉시 chaincode 호출로 이동할 수 있습니다.

터미널 2 - 체인 코드 빌드와 시작(Terminal 2 - Build & start the chaincode) 
----------------------------------------

.. code:: bash

  docker exec -it chaincode bash

다음이 나와야 합니다.

.. code:: bash

  root@d2629980e76b:/opt/gopath/src/chaincode#

이제, 여러분의 chaincode를 컴파일합니다.

.. code:: bash

  cd chaincode_example02
  go build

이제 chaincode를 실행합니다.

.. code:: bash

  CORE_PEER_ADDRESS=peer:7051 CORE_CHAINCODE_ID_NAME=mycc:0 ./chaincode_example02

체인 코드는 피어와 체인 코드 로그로 시작되어 피어와의 성공적인 등록을 나타냅니다. 이 단계에서 체인 코드는 어떤 채널과도 연결되지 않습니다. 이는 instantiate 명령을 사용하여 후속 단계에서 수행됩니다.

터미널 3 - 체인 코드 사용(Terminal 3 - Use the chaincode) 
------------------------------

--peer-chaincodedev 모드에 있더라도 수명 코드 시스템 체인 코드가 정상적으로 검사할 수 있도록 체인 코드를 설치해야합니다. 이 요구 사항은 나중에 --peer-chaincodedevmode에서 제거 될 수 있습니다.

CLI 컨테이너를 활용하여 이러한 호출을 유도합니다.

.. code:: bash

  docker exec -it cli bash

.. code:: bash

  peer chaincode install -p chaincodedev/chaincode/chaincode_example02 -n mycc -v 0
  peer chaincode instantiate -n mycc -v 0 -c '{"Args":["init","a","100","b","200"]}' -C myc

이제 invoke를 실행하여 a에서 b로 10을 이동하십시오.

.. code:: bash

  peer chaincode invoke -n mycc -c '{"Args":["invoke","a","b","10"]}' -C myc

마지막으로 a에 쿼리 요청을 합니다. 값 90이 나와야 합니다.

.. code:: bash

  peer chaincode query -n mycc -c '{"Args":["query","a"]}' -C myc

새로운 체인 코드 테스트(Testing new chaincode) 
---------------------

기본적으로 우리는 ``chaincode_example02``를 마운트만합니다 . 그러나 다른 체인 코드를 ``chaincode`` 하위 디렉토리 에 추가하고 네트워크를 다시 시작하여 쉽게 다른 체인 코드를 테스트할 수 있습니다 . 이 시점에서 그들은 당신의 ``chaincode``컨테이너에서 접근 가능할 것 입니다.

your network.  At this point they will be accessible in your ``chaincode`` container.

.. Licensed under Creative Commons Attribution 4.0 International License
     https://creativecommons.org/licenses/by/4.0/
