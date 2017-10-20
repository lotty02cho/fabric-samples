# 높은 처리량 네트워크(High-Throughput Network)

## 목적(Purpose)
This network is used to understand how to properly design the chaincode data model when handling thousands of transactions per second which all
update the same asset in the ledger. A naive implementation would use a single key to represent the data for the asset, and the chaincode would
then attempt to update this key every time a transaction involving it comes in. However, when many transactions all come in at once, in the time
between when the transaction is simulated on the peer (i.e. read-set is created) and it's ready to be committed to the ledger, another transaction
may have already updated the same value. Thus, in the simple implementation, the read-set version will no longer match the version in the orderer,
and a large number of parallel transactions will fail. To solve this issue, the frequently updated value is instead stored as a series of deltas
which are aggregated when the value must be retrieved. In this way, no single row is frequently read and updated, but rather a collection of rows
is considered.
이 네트워크는 초당 수천 건의 트랜잭션을 처리 할 때 체인 코드 데이터 모델을 올바르게 설계하는 방법을 이해하는 데 사용됩니다.이 트랜잭션은 모두 원장의 동일한 자산을 업데이트합니다. 이 간단한 구현은 단일 키를 사용하여 자산의 데이터를 나타낼 수 있으며, 체인 코드는 관련된 키가 들어올 때마다이 키를 업데이트하려고 시도합니다. 그러나 많은 트랜잭션이 모두 동시에 들어올 때, 트랜잭션이 피어(peer)에서 시뮬레이트 될 때(즉, 읽기 - 집합이 생성 될 때) 장부에 커밋 될 준비가되면 다른 트랜잭션이 이미 동일한 값을 업데이트했을 수 있습니다. 따라서 간단한 구현에서 읽기 세트 버전은 더 이상 주문자의 버전과 일치하지 않으며 많은 수의 병렬 트랜잭션이 실패합니다. 이 문제를 해결하기 위해, 자주 갱신되는 값은 대신 값을 검색해야 할 때 집계되는 일련의 델타로 저장됩니다. 이런 식으로 한 행을 자주 읽고 업데이트하는 것이 아니라 행 모음을 고려합니다.

## 유스 케이스(Use Case)
The primary use case for this chaincode data model design is for applications in which a particular asset has an associated amount that is
frequently added to or removed from. For example, with a bank or credit card account, money is either paid to or paid out of it, and the amount
of money in the account is the result of all of these additions and subtractions aggregated together. A typical person's bank account may not be
used frequently enough to require highly-parallel throughput, but an organizational account used to store the money collected from customers on an
e-commerce platform may very well receive a very high number of transactions from all over the world all at once. In fact, this use case is the only
use case for crypto currencies like Bitcoin: a user's unspent transaction output (UTXO) is the result of all transactions he or she has been a part of since joining the blockchain. Other use cases that can employ this technique might be IOT sensors which frequently update their sensed value in the cloud.

By adopting this method of storing data, an organization can optimize their chaincode to store and record transactions as quickly as possible and can aggregate ledger records into one value at the time of their choosing without sacrificing transaction performance. Given the state-machine design of Hyperledger Fabric, however, careful considerations need to be given to the data model design for the chaincode.

Let's look at some concrete use cases and how an organization might implement high-throughput storage. These cases will try and explore some of the
advantages and disadvantages of such a system, and how to overcome them.

이 체인 코드 데이터 모델 설계의 기본 유스 케이스는 특정 자산에 관련 금액이있는 애플리케이션의 경우입니다. 자주 추가되거나 제거되었습니다.
예를 들어, 은행이나 신용 카드 계좌에서는 돈을 지불하거나 돈을 지불하고, 계좌에있는 금액은 함께 합산 한 모든 추가 및 빼기의 결과입니다.
일반적인 사람의 은행 계좌는 병렬 처리량이 많이 필요할 정도로 자주 사용되지 않을 수 있지만 전자 상거래 플랫폼에 고객으로부터 수집 된 금액을 저장하는 데 사용되는 조직 계정은 전체에서 전세계적으로 한꺼번에 매우 많은 거래를 수신 할 수 있습니다. 사실 이 사용 사례는 Bitcoin과 같은 암호화 통화의 유일한 사용 사례입니다. 사용자의 미사용 트랜잭션 출력 (UTXO)은 블록 체인에 참가한 이후로 참여한 모든 트랜잭션의 결과입니다. 이 기술을 사용할 수있는 다른 사용 사례는 클라우드에서 감지 된 값을 자주 업데이트하는 IOT 센서 일 수 있습니다.

조직은 데이터를 저장하는이 방법을 채택함으로써 가능한 한 빨리 트랜잭션을 저장하고 기록 할 수 있도록 체인 코드를 최적화 할 수 있습니다. 트랜잭션 성능을 희생시키지 않으면 서 장부 원 레코드를 선택시 한 값으로 집계합니다. 그러나 Hyperledger Fabric의 상태 - 기계 설계를 고려할 때, 체인 코드에 대한 데이터 모델 설계에 신중한 고려 사항이 주어져야합니다.

구체적인 사용 사례와 조직에서 처리량이 많은 저장소를 구현하는 방법을 살펴 보겠습니다. 이러한 경우는 그러한 시스템의 장점과 단점 중 일부를 검색하고 이를 극복하는 방법을 시도합니다.

#### 사례 1 (IOT) : 박서 건설 분석가(Example 1 (IOT): Boxer Construction Analysts)

Boxer Construction Analysts is an IOT company focused on enabling real-time monitoring of large, expensive assets (machinery) on commercial
construction projects. They've partnered with the only construction vehicle company in New York, Condor Machines Inc., to provide a reliable,
auditable, and replayable monitoring system on their machines. This allows Condor to monitor their machines and address problems as soon as
they occur while providing end-users with a transparent report on machine health, which helps keep the customers satisfied.

The vehicles are outfitted with many sensors each of which broadcasts updated values at frequencies ranging from several times a second to
several times a minute. Boxer initially sets up their chaincode so that the central machine computer pushes these values out to the blockchain
as soon as they're produced, and each sensor has its own row in the ledger which is updated when a new value comes in. While they find that
this works fine for the sensors which only update several times a minute, they run into some issues when updating the faster sensors. Often,
the blockchain skips several sensor readings before adding a new one, defeating the purpose of having a fast, always-on sensor. The issue they're
running into is that they're sending update transactions so fast that the version of the row is changed between the creation of a transaction's
read-set and committing that transaction to the ledger. The result is that while a transaction is in the process of being committed, all future
transactions are rejected until the commitment process is complete and a new, much later reading updates the ledger.

To address this issue, they adopt a high-throughput design for the chaincode data model instead. Each sensor has a key which identifies it within the
ledger, and the difference between the previous reading and the current reading is published as a transaction. For example, if a sensor is monitoring
engine temperature, rather than sending the following list: 220F, 223F, 233F, 227F, the sensor would send: +220, +3, +10, -6 (the sensor is assumed
to start a 0 on initialization). This solves the throughput problem, as the machine can post delta transactions as fast as it wants and they will all
eventually be committed to the ledger in the order they were received. Additionally, these transactions can be processed as they appear in the ledger
by a dashboard to provide live monitoring data. The only difference the engineers have to pay attention to in this case is to make sure the sensors can send deltas from the previous reading, rather than fixed readings.

Boxer Construction Analysts는 상업용 건설 프로젝트에서 크고 값 비싼 자산 (기계류)을 실시간으로 모니터링 할 수 있도록하는 IOT 회사입니다. 뉴욕에있는 유일한 건설 차량 회사 인 Condor Machines Inc.와 파트너 관계를 맺어 안정적이고 감사 가능하며 재생 가능한 모니터링 시스템을 제공합니다. 이를 통해 Condor는 컴퓨터를 모니터링하고 문제가 발생하자마자 해결할 수 있으며 최종 사용자에게 장비 상태에 대한 투명한 보고서를 제공함으로써 고객 만족을 유지할 수 있습니다.

차량에는 여러 센서가 장착되어 있으며, 각 센서는 분당 몇 초에서 수 시간에 이르는 주파수에서 업데이트 된 값을 방송합니다. Boxer는 초기에 체인 코드를 설정하여 중앙 컴퓨터 컴퓨터가 생성 된 값을 즉시 블록 체인에 푸시하고 각 센서에는 원장에 새로운 값이 들어올 때 업데이트되는 자체 행이 있습니다. 분당 몇 번만 업데이트하는 센서에서는이 기능이 제대로 작동하므로 빠른 센서를 업데이트 할 때 문제가 발생합니다. 종종 블록 체인은 새로운 센서를 판독하기 전에 여러 센서 판독 값을 건너 뛰고 고속의 상시 센서를 사용하는 목적을 무효화합니다. 그들이 실행중인 문제는 업데이트 트랜잭션을 너무 빨리 보내 트랜잭션의 읽기 집합을 생성하고 해당 트랜잭션을 장부에 커밋하는 과정에서 행의 버전이 변경된다는 것입니다. 결과적으로 트랜잭션이 커밋되는 동안 모든 커밋 된 트랜잭션은 커밋 프로세스가 완료 될 때까지 거부되고 새롭고 훨씬 나중에 읽는 트랜잭션이 장부를 업데이트합니다.

이 문제를 해결하기 위해 체인 코드 데이터 모델 대신 고효율 설계를 채택했습니다. 각 센서에는 원장 내에서이를 식별하는 키가 있으며 이전 판독 값과 현재 판독 값의 차이가 트랜잭션으로 게시됩니다. 예를 들어, 센서가 220F, 223F, 233F, 227F와 같은 목록을 보내지 않고 엔진 온도를 모니터링하는 경우 센서는 다음을 전송합니다 : +220, +3, + 10, -6 (센서는 초기화시 0). 이것은 기계가 델타 트랜잭션을 원하는만큼 빠르게 게시 할 수 있기 때문에 처리량 문제를 해결하며, 결국 모든 원가가 접수 된 순서대로 원장에게 위탁됩니다. 또한 실시간 모니터링 데이터를 제공하기 위해 대시 보드에서 대부에 표시되는대로 이러한 트랜잭션을 처리 할 수 ​​있습니다. 이 경우 엔지니어가주의해야 할 유일한 차이점은 센서가 고정 된 판독 값이 아닌 이전 판독 값에서 델타를 전송할 수 있는지 확인하는 것입니다.

#### 예제 2 (잔액 이체) : Robinson Credit Co.(Example 2 (Balance Transfer): Robinson Credit Co.)

Robinson Credit Co. provides credit and financial services to large businesses. As such, their accounts are large, complex, and accessed by many
people at once at any time of the day. They want to switch to blockchain, but are having trouble keeping up with the number of deposits and
withdrawals happening at once on the same account. Additionally, they need to ensure users never withdraw more money than is available
on an account, and transactions that do get rejected. The first problem is easy to solve, the second is more nuanced and requires a variety of
strategies to accommodate high-throughput storage model design.

To solve throughput, this new storage model is leveraged to allow every user performing transactions against the account to make that transaction in terms
of a delta. For example, global e-commerce company America Inc. must be able to accept thousands of transactions an hour in order to keep up with
their customer's demands. Rather than attempt to update a single row with the total amount of money in America Inc's account, Robinson Credit Co.
accepts each transaction as an additive delta to America Inc's account. At the end of the day, America Inc's accounting department can quickly
retrieve the total value in the account when the sums are aggregated.

However, what happens when American Inc. now wants to pay its suppliers out of the same account, or a different account also on the blockchain?
Robinson Credit Co. would like to be assured that America Inc.'s accounting department can't simply overdraw their account, which is difficult to
do while at the same enabling transactions to happen quickly, as deltas are added to the ledger without any sort of bounds checking on the final
aggregate value. There are a variety of solutions which can be used in combination to address this.

Solution 1 involves polling the aggregate value regularly. This happens separate from any delta transaction, and can be performed by a monitoring
service setup by Robinson themselves so that they can at least be guaranteed that if an overdraw does occur, they can detect it within a known
number of seconds and respond to it appropriately (e.g. by temporarily shutting off transactions on that account), all of which can be automated.
Furthermore, thanks to the decentralized nature of Fabric, this operation can be performed on a peer dedicated to this function that would not
slow down or impact the performance of peers processing customer transactions.

Solution 2 involves breaking up the submission and verification steps of the balance transfer. Balance transfer submissions happen very quickly
and don't bother with checking overdrawing. However, a secondary process reviews each transaction sent to the chain and keeps a running total,
verifying that none of them overdraw the account, or at the very least that aggregated withdrawals vs deposits balance out at the end of the day.
Similar to Solution 1, this system would run separate from any transaction processing hardware and would not incur a performance hit on the
customer-facing chain.

Solution 3 involves individually tailoring the smart contracts between Robinson and America Inc, leveraging the power of chaincode to customize
spending limits based on solvency proofs. Perhaps a limit is set on withdrawal transactions such that anything below \$1000 is automatically processed
and assumed to be correct and at minimal risk to either company simply due to America Inc. having proved solvency. However, withdrawals above \$1000
must be verified before approval and admittance to the chain.

Robinson Credit Co.는 대기업에 신용 및 금융 서비스를 제공합니다. 따라서 계정은 크고 복잡하며 많은 사람들이 하루 중 언제든지 즉시 액세스 할 수 있습니다. 그들은 블록 체인 (blockchain)으로 전환하고 싶지만 같은 계좌에서 한번에 예금과 인출 ​​횟수를 따라 잡는 데 어려움을 겪고 있습니다. 또한 사용자는 계정에서 사용할 수있는 것보다 많은 돈을 인출하지 않도록하고 거부되는 거래를 방지해야합니다. 첫 번째 문제는 해결하기 쉽고 두 번째 문제는 더 미묘한 차이가 있으며 고효율 스토리지 모델 설계를 수용하기위한 다양한 전략이 필요합니다.

처리량을 해결하기 위해이 새로운 스토리지 모델을 활용하여 계정에 대한 트랜잭션을 수행하는 모든 사용자가 델타 단위로 트랜잭션을 수행 할 수 있습니다. 예를 들어 글로벌 전자 상거래 회사 인 America Inc.는 고객의 요구에 부응하기 위해 시간당 수천 건의 거래를 수락 할 수 있어야합니다. Robinson Credit Co.는 America Inc의 계정에있는 총 금액으로 단일 행을 업데이트하려고 시도하지 않고 각 거래를 America Inc의 계정에 대한 추가 델타로 허용합니다. 하루가 끝나면 America Inc의 회계 부서는 합계가 합산되면 계정의 총 가치를 빠르게 검색 할 수 있습니다.

그러나 American Inc.가 현재 공급 업체에게 동일한 계정 또는 블록 체인의 다른 계정으로 지불하기를 원하면 어떻게됩니까? Robinson Credit Co.는 America Inc.의 회계 부서가 자신의 계좌를 단순히 과도하게 인수 처리 할 수는 없으므로 거래를 신속하게 처리 할 수있는 반면 델타는 원장에 아무런 정렬없이 추가되므로 최종 집계 값 확인 범위. 이 문제를 해결하기 위해 여러 가지 솔루션을 조합하여 사용할 수 있습니다.

해결 방법 1은 집계 값을 정기적으로 폴링하는 것입니다. 이것은 델타 트랜잭션과 별개로 발생하며 Robinson 자체의 모니터링 서비스 설정으로 수행 할 수 있으므로 적어도 초과 실행이 발생하면 알려진 시간 내에이를 감지하고 적절히 대응할 수 있습니다 (예 : 해당 계정의 트랜잭션을 일시적으로 종료하여) 자동화 할 수 있습니다. 또한 Fabric의 분산 된 특성 덕분에이 작업은 고객 트랜잭션을 처리하는 피어의 성능을 저하 시키거나 성능에 영향을 미치지 않는이 기능 전용의 피어에서 수행 할 수 있습니다.

해결 방법 2는 잔액 이체의 제출 및 확인 단계를 해체하는 것입니다. 잔액 이체 제출은 매우 신속하게 이루어지며 초과 인출 확인에 신경 쓸 필요가 없습니다. 그러나 2 차 프로세스는 체인으로 전송 된 각 트랜잭션을 검토하고 누적 금액을 유지하여 계정을 초과 작성하지 않았는지 확인합니다. 또는 최소한 집계 된 금액과 예금액은 하루가 끝날 때 균형을 이룹니다. 솔루션 1과 마찬가지로이 시스템은 모든 트랜잭션 처리 하드웨어와 분리되어 실행되며 고객 대면 체인에서 성능이 저하되지 않습니다.

솔루션 3은 로빈슨과 아메리카 사이의 스마트 계약을 개별적으로 조정하여 체인 코드의 힘을 바탕으로 지급 능력 증명을 기준으로 지출 한도를 맞춤화합니다. 아마도 한도액은 $ 1000 이하의 금액이 자동으로 처리되고 올바른 것으로 가정하고 지불 능력이 입증 된 America Inc.만으로도 어느 회사에 대해서도 최소한의 위험 부담으로 철수 거래에 설정됩니다. 그러나 체인에 대한 승인 및 승인 전에 $ 1000 이상의 인출을 확인해야합니다.

## 방법(How)
This sample provides the chaincode and scripts required to run a high-throughput application. For ease of use, it runs on the same network which is brought
up by `byfn.sh` in the `first-network` folder within `fabric-samples`, albeit with a few small modifications. The instructions to build the network
and run some invocations are provided below.
(이 샘플은 처리량이 많은 응용 프로그램을 실행하는 데 필요한 체인 코드와 스크립트를 제공합니다. 사용의 용이성을 위해 동일한 네트워크에서 실행됩니다.이 네트워크는 `fabric-samples` 내의 `first-network` 폴더에있는 `byfn.sh`에 의해 약간의 수정이 이루어졌지만 실행됩니다. 네트워크를 구축하고 호출을 실행하는 방법은 아래에 나와 있습니다.)

### 네트워크 구축하기(Build your network)
1. `cd` into the `first-network` folder within `fabric-samples`, e.g. `cd ~/fabric-samples/first-network`
   (cd명령어를 이용하여 `fabric-sample` 안에 `first-network` 폴더로 이동합니다.)
2. Open `docker-compose-cli.yaml` in your favorite editor, and edit the following lines:
   (자주 쓰는 편집기(Atom, VScode)에서 `docker-composer-cli.yaml` 파일을 열고, 아래의 라인을 수정합니다.)
  * In the `volumes` section of the `cli` container, edit the second line which refers to the chaincode folder to point to the chaincode folder
    within the `high-throughput` folder, e.g.
    (`cli` 컨테이너의`volumes` 섹션에서,`high-throughput` 폴더 내의 chaincode 폴더를 가리키도록 chaincode 폴더를 가리키는 두 번째 라인을 편집하십시오.)


    `./../chaincode/:/opt/gopath/src/github.com/hyperledger/fabric/examples/chaincode/go` --> 
    `./../high-throughput/chaincode/:/opt/gopath/src/github.com/hyperledger/fabric/examples/chaincode/go`
  * Again in the `volumes` section, edit the fourth line which refers to the scripts folder so it points to the scripts folder within the
    `high-throughput` folder, e.g.
    (다시 `volumes` 섹션에서 스크립트 폴더를 가리키는 네 번째 줄을 편집하여 `high-throughput` 폴더 내의 scripts 폴더를 가리킵니다.)

    `./scripts:/opt/gopath/src/github.com/hyperledger/fabric/peer/scripts/` --> 
    `./../high-throughput/scripts/:/opt/gopath/src/github.com/hyperledger/fabric/peer/scripts/`

  * Finally, comment out the `command` section by placing a `#` before it, e.g.
    (마지막으로 `command` 섹션 앞에`#`기호를 붙여 주석 처리하십시오.)
    
    `#command: /bin/bash -c './scripts/script.sh ${CHANNEL_NAME}; sleep $TIMEOUT'`

3. We can now bring our network up by typing in `./byfn.sh -m up -c mychannel`
   (이제 우리는`./byfn.sh -m up -c mychannel`을 타이핑하여 네트워크를 가동시킬 수 있습니다.)
4. Open a new terminal window and enter the CLI container using `docker exec -it cli bash`, all operations on the network will happen within
   this container from now on.
   (새 터미널 창을 열고`docker exec -it cli bash`를 사용하여 CLI 컨테이너에 들어가면 네트워크의 모든 작업이 이제부터이 컨테이너에서 수행됩니다.)

### 체인 코드 설치 및 인스턴스화(Install and instantiate the chaincode)
1. Once you're in the CLI container run `cd scripts` to enter the `scripts` folder
   (CLI 컨테이너에 들어가면`cd scripts`를 실행하여`scripts` 폴더에 들어갑니다.)
2. Set-up the environment variables by running `source setclienv.sh`
   (`source setclienv.sh`를 실행하여 환경 변수를 설정하십시오.)
3. Set-up your channels and anchor peers by running `./channel-setup.sh`
   (`. / channel-setup.sh`을 실행하여 채널과 앵커 피어를 설정하십시오.)
4. Install your chaincode by running `./install-chaincode.sh 1.0`. The only argument is a number representing the chaincode version, every time
   you want to install and upgrade to a new chaincode version simply increment this value by 1 when running the command, e.g. `./install-chaincode.sh 2.0`
   (./install-chaincode.sh 1.0을 실행하여 체인 코드를 설치하십시오. 유일한 인수는 매번 체인 코드 버전을 나타내는 숫자입니다.
    새로운 체인 코드 버전을 설치하고 업그레이드 하려면 이 명령을 실행할 때이 값을 1 씩 증가 시키면됩니다.)
   `./install-chaincode.sh 2.0`
5. Instantiate your chaincode by running `./instantiate-chaincode.sh 1.0`. The version argument serves the same purpose as in `./install-chaincode.sh 1.0`
   and should match the version of the chaincode you just installed. In the future, when upgrading the chaincode to a newer version,
   `./upgrade-chaincode.sh 2.0` should be used instead of `./instantiate-chaincode.sh 1.0`.
   (`./instantiate-chaincode.sh 1.0`을 실행하여 체인 코드를 인스턴스화하십시오. version 인자는`./install-chaincode.sh 1.0`과 같은 목적을합니다.
    방금 설치 한 체인 코드의 버전과 일치해야합니다. 앞으로 체인 코드를 새로운 버전으로 업그레이드 할 때,
    `./instantiate-chaincode.sh 1.0 '대신`./upgrade-chaincode.sh 2.0`을 사용해야합니다.)
6. Your chaincode is now installed and ready to receive invocations
   (체인 코드가 설치되었으며 호출을 받을 준비가되었습니다.)

### 체인코드 호출(Invoke the chaincode)
All invocations are provided as scripts in `scripts` folder; these are detailed below.
(모든 호출은`scripts` 폴더에 스크립트로 제공됩니다; 이것들은 아래에 자세히 나와있습니다.)

#### Update
The format for update is: `./update-invoke.sh name value operation` where `name` is the name of the variable to update, `value` is the value to
add to the variable, and `operation` is either `+` or `-` depending on what type of operation you'd like to add to the variable. In the future,
multiply/divide operations will be supported (or add them yourself to the chaincode as an exercise!)
(업데이트 형식은 다음과 같습니다.`./update-invoke.sh name value operation``name`은 업데이트 할 변수의 이름이고, 'value`는 값입니다.
변수에 추가하십시오.`operation`은 변수에 추가 할 조작 유형에 따라`+`또는`-`입니다. 앞으로, 곱하기 / 나누기 연산이 지원됩니다. (또는 연습으로 체인 코드에 직접 추가하십시오!))

Example: `./update-invoke.sh myvar 100 +`

#### Get
The format for get is: `./get-invoke.sh name` where `name` is the name of the variable to get.
(get의 형식은`./get-invoke.sh name`입니다. 여기서`name`은 얻을 변수의 이름입니다.)

Example: `./get-invoke.sh myvar`

#### Delete
The format for delete is: `./delete-invoke.sh name` where `name` is the name of the variable to delete.
(delete의 형식은`./delete-invoke.sh name`입니다. 여기서`name`은 삭제할 변수의 이름입니다.)

Example: `./delete-invoke.sh myvar`

#### Prune
Pruning takes all the deltas generated for a variable and combines them all into a single row, deleting all previous rows. This helps cleanup
the ledger when many updates have been performed. There are two types of pruning: `prunefast` and `prunesafe`. Prune fast performs the deletion
and aggregation simultaneously, so if an error happens along the way data integrity is not guaranteed. Prune safe performs the aggregation first,
backs up the results, then performs the deletion. This way, if an error occurs along the way, data integrity is maintained.

The format for pruning is: `./[prunesafe|prunefast]-invoke.sh name` where `name` is the name of the variable to prune.

(가지 치기는 변수에 대해 생성 된 모든 델타를 취해 이들을 모두 단일 행으로 결합하여 이전 행을 모두 삭제합니다. 이렇게하면 많은 업데이트가 수행되었을 때 원장 정리에 도움이됩니다. 가지 치기에는 'prunefast'와 'prunesafe'의 두 가지 유형이 있습니다. Prune Fast는 삭제 및 집계를 동시에 수행하므로 데이터 무결성이 보장되지 않는 방식으로 오류가 발생하는 경우 프룬 안전은 먼저 집계를 수행하고 결과를 백업 한 다음 삭제를 수행합니다. 이런 식으로 오류가 발생하면 데이터 무결성이 유지됩니다.

프룬의 형식은 다음과 같습니다.`./[prunesafe|prunefast]-invoke.sh name` 여기서 `name`은 프룬(prune)할 변수의 이름입니다.)

Example: `./prunefast-invoke.sh myvar` or `./prunesafe-invoke.sh myvar`

### 네트워크 테스트(Test the Network)
Two scripts are provided to show the advantage of using this system when running many parallel transactions at once: `many-updates.sh` and
`many-updates-traditional.sh`. The first script accepts the same arguments as `update-invoke.sh` but duplicates the invocation 1000 times
and in parallel. The final value, therefore, should be the given update value * 1000. Run this script to confirm that your network is functioning
properly. You can confirm this by checking your peer and orderer logs and verifying that no invocations are rejected due to improper versions.

The second script, `many-updates-traditional.sh`, also sends 1000 transactions but using the traditional storage system. It'll update a single
row in the ledger 1000 times, with a value incrementing by one each time (i.e. the first invocation sets it to 0 and the last to 1000). The
expectation would be that the final value of the row is 999. However, the final value changes each time this script is run and you'll find
errors in the peer and orderer logs.

There is one other script, `get-traditional.sh`, which simply gets the value of a row in the traditional way, with no deltas.

(`many-updates.sh`와`many-updates-traditional.sh`와 같이 많은 병렬 트랜잭션을 동시에 실행할 때이 시스템을 사용하는 이점을 보여주기 위해 두 개의 스크립트가 제공됩니다. 첫 번째 스크립트는`update-invoke.sh`와 동일한 인수를 받아들이지 만 호출을 1000 번 동시에 병렬로 복제합니다. 따라서 최종 값은 제공된 업데이트 값 * 1000이어야합니다.이 스크립트를 실행하여 네트워크가 올바르게 작동하는지 확인하십시오. 피어 및 주문자 로그를 확인하고 부적절한 버전으로 인해 호출이 거부되지 않았는지 확인하여이를 확인할 수 있습니다.

두 번째 스크립트 `many-updates-traditional.sh`는 또한 1000 개의 트랜잭션을 보내지 만 전통적인 저장소 시스템을 사용합니다. 원장의 행 하나를 1000 번 업데이트 할 때마다 값이 1 씩 증가합니다(즉, 첫 번째 호출은 0으로 설정되고 마지막으로 1000으로 설정 됨). 예상되는 행의 최종 값은 999입니다. 그러나이 스크립트가 실행될 때마다 최종 값이 변경되고 피어 및 순서자 로그에 오류가 있습니다.

또 다른 스크립트인 `get-traditional.sh`가 있습니다. 델타없이 전통적인 방식으로 행의 가치를 얻는 것입니다.)

Examples:
`./many-updates.sh testvar 100 +` --> final value from `./get-invoke.sh` should be 100000
`./many-updates-traditional.sh testvar` --> final value from `./get-traditional.sh testvar` is undefined
