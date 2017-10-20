## 잔액 이체(Balance transfer)

A sample Node.js app to demonstrate **__fabric-client__** & **__fabric-ca-client__** Node.js SDK APIs
**__fabric-client__** 와 **__fabric-ca-client__**의 Node.js SDK API를 보여주는 예제 Node.js 앱

### 전제 조건 및 설정(Prerequisites and setup):

* 도커 - v1.12 이상([Docker](https://www.docker.com/products/overview) - v1.12 or higher)
* docker compose - v1.8 이상([Docker Compose](https://docs.docker.com/compose/overview/) - v1.8 or higher)
* [Git client](https://git-scm.com/downloads) - 복제 명령에 필요(needed for clone commands)
* **Node.js** v6.9.0 - 6.10.0 (( 노드 v7 +는 지원되지 않음 ) __Node v7+ is not supported__ )
* [Download Docker images](http://hyperledger-fabric.readthedocs.io/en/latest/samples.html#binaries)

```
cd fabric-samples/balance-transfer/
```

위의 설정을 완료하면 다음과 같은 도커 컨테이너 구성으로 로컬 네트워크를 프로비저닝할 수 있습니다.

* 2 CAs
* A SOLO orderer
* 4 peers (2 peers per Org)

#### 유물(Artifacts) 
* Crypto material has been generated using the **cryptogen** tool from Hyperledger Fabric and mounted to all peers, the orderering node and CA containers. More details regarding the cryptogen tool are available [here](http://hyperledger-fabric.readthedocs.io/en/latest/build_network.html#crypto-generator).
Crypto 자료는 Hyperledger Fabric의 cryptogen 도구를 사용하여 생성되었으며 모든 피어, orderering 노드와 CA 컨테이너에 마운트 됩니다. cryptogen 도구에 대한 자세한 내용은 여기를 참조하십시오.
* An Orderer genesis block (genesis.block) and channel configuration transaction (mychannel.tx) has been pre generated using the **configtxgen** tool from Hyperledger Fabric and placed within the artifacts folder. More details regarding the configtxgen tool are available [here](http://hyperledger-fabric.readthedocs.io/en/latest/build_network.html#configuration-transaction-generator).
Crypto 자료는 Hyperledger Fabric의 cryptogen 도구를 사용하여 생성되었으며 모든 피어, orderering 노드와 CA 컨테이너에 마운트 됩니다. cryptogen 도구에 대한 자세한 내용은 여기를 참조하십시오.

## 샘플 프로그램 실행하기(Running the sample program) 

잔액 이체 샘플을 실행하는 데 사용할 수 있는 두 가지 옵션이 있습니다.

### 옵션 1(Option 1): 

##### 터미널 창 1(Terminal Window 1) 

* docker-compose를 사용하여 네트워크 실행

```
docker-compose -f artifacts/docker-compose.yaml up
```
##### 터미널 창 2(Terminal Window 2) 

* abric-client 및 fabric-ca-client 노드 모듈 설치

```
npm install
```

* PORT 4000에서 노드 응용 프로그램 시작

```
PORT=4000 node app
```

##### 터미널 창 3(Terminal Window 3) 

* Execute the REST APIs from the section [Sample REST APIs Requests](https://github.com/hyperledger/fabric-samples/tree/master/balance-transfer#sample-rest-apis-requests)
아래에 샘플 REST API 요청 섹션에 있는 REST API를 실행하십시오.

### 옵션 2(Option 2): 

##### 터미널 창 1(Terminal Window 1) 

```
cd fabric-samples/balance-transfer

./runApp.sh

```

* 로컬 컴퓨터에서 필요한 네트워크를 시작합니다.
* fabric-client 및 fabric-ca-client 노드 모듈을 설치합니다.
* 그리고 PORT 4000에서 노드 앱을 시작합니다.


##### 터미널 창 2(Terminal Window 2) 


다음 쉘 스크립트가 JSON을 제대로 파싱하려면``jq``를 설치해야합니다 :

지침 [https://stedolan.github.io/jq/](https://stedolan.github.io/jq/)

어플리케이션이 터미널 1에서 시작되면, 다음 스크립트를 실행하여 API를 테스트하십시오 - **testAPIs.sh**:
```
cd fabric-samples/balance-transfer

./testAPIs.sh

```

## 샘플 REST API 요청(Sample REST APIs Requests) 

### 로그인 요청(Login Request) 

* Organization - Org1에 새로운 사용자 등록 및 등록 - **Org1**:

`curl -s -X POST http://localhost:4000/users -H "content-type: application/x-www-form-urlencoded" -d 'username=Jim&orgName=org1'`

**결과:**

```
{
  "success": true,
  "secret": "RaxhMgevgJcm",
  "message": "Jim enrolled Successfully",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0OTQ4NjU1OTEsInVzZXJuYW1lIjoiSmltIiwib3JnTmFtZSI6Im9yZzEiLCJpYXQiOjE0OTQ4NjE5OTF9.yWaJhFDuTvMQRaZIqg20Is5t-JJ_1BP58yrNLOKxtNI"
}
```

The response contains the success/failure status, an **enrollment Secret** and a **JSON Web Token (JWT)** that is a required string in the Request Headers for subsequent requests.
응답에는 성공/실패 상태, 등록 비밀과 후속 요청에 대한 요청 헤더의 필수 문자열인 JSON Web Token(JWT)이 포함됩니다.

### 채널 요청 만들기(Create Channel request) 

```
curl -s -X POST \
  http://localhost:4000/channels \
  -H "authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0OTQ4NjU1OTEsInVzZXJuYW1lIjoiSmltIiwib3JnTmFtZSI6Im9yZzEiLCJpYXQiOjE0OTQ4NjE5OTF9.yWaJhFDuTvMQRaZIqg20Is5t-JJ_1BP58yrNLOKxtNI" \
  -H "content-type: application/json" \
  -d '{
	"channelName":"mychannel",
	"channelConfigPath":"../artifacts/channel/mychannel.tx"
}'
```

헤더 **authorization** 에는 `POST /users` 호출에서 반환된 JWT가 포함되어야합니다.

### 채널 가입 요청(Join Channel request) 

```
curl -s -X POST \
  http://localhost:4000/channels/mychannel/peers \
  -H "authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0OTQ4NjU1OTEsInVzZXJuYW1lIjoiSmltIiwib3JnTmFtZSI6Im9yZzEiLCJpYXQiOjE0OTQ4NjE5OTF9.yWaJhFDuTvMQRaZIqg20Is5t-JJ_1BP58yrNLOKxtNI" \
  -H "content-type: application/json" \
  -d '{
	"peers": ["peer1","peer2"]
}'
```
### 체인코드 설치(Install chaincode) 

```
curl -s -X POST \
  http://localhost:4000/chaincodes \
  -H "authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0OTQ4NjU1OTEsInVzZXJuYW1lIjoiSmltIiwib3JnTmFtZSI6Im9yZzEiLCJpYXQiOjE0OTQ4NjE5OTF9.yWaJhFDuTvMQRaZIqg20Is5t-JJ_1BP58yrNLOKxtNI" \
  -H "content-type: application/json" \
  -d '{
	"peers": ["peer1","peer2"],
	"chaincodeName":"mycc",
	"chaincodePath":"github.com/example_cc",
	"chaincodeVersion":"v0"
}'
```

### 체인 코드 인스턴스화(Instantiate chaincode) 

```
curl -s -X POST \
  http://localhost:4000/channels/mychannel/chaincodes \
  -H "authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0OTQ4NjU1OTEsInVzZXJuYW1lIjoiSmltIiwib3JnTmFtZSI6Im9yZzEiLCJpYXQiOjE0OTQ4NjE5OTF9.yWaJhFDuTvMQRaZIqg20Is5t-JJ_1BP58yrNLOKxtNI" \
  -H "content-type: application/json" \
  -d '{
	"chaincodeName":"mycc",
	"chaincodeVersion":"v0",
	"args":["a","100","b","200"]
}'
```

### 요청 호출(Invoke request)

```
curl -s -X POST \
  http://localhost:4000/channels/mychannel/chaincodes/mycc \
  -H "authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0OTQ4NjU1OTEsInVzZXJuYW1lIjoiSmltIiwib3JnTmFtZSI6Im9yZzEiLCJpYXQiOjE0OTQ4NjE5OTF9.yWaJhFDuTvMQRaZIqg20Is5t-JJ_1BP58yrNLOKxtNI" \
  -H "content-type: application/json" \
  -d '{
	"fcn":"move",
	"args":["a","b","10"]
}'
```
**참고:** 후속 쿼리 트랜잭션에서이 문자열을 전달하려면 응답에서 트랜잭션 ID를 저장해야합니다.

### 체인 코드 쿼리(Chaincode Query) 

```
curl -s -X GET \
  "http://localhost:4000/channels/mychannel/chaincodes/mycc?peer=peer1&fcn=query&args=%5B%22a%22%5D" \
  -H "authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0OTQ4NjU1OTEsInVzZXJuYW1lIjoiSmltIiwib3JnTmFtZSI6Im9yZzEiLCJpYXQiOjE0OTQ4NjE5OTF9.yWaJhFDuTvMQRaZIqg20Is5t-JJ_1BP58yrNLOKxtNI" \
  -H "content-type: application/json"
```

### BlockNumber로 블록 쿼리하기(Query Block by BlockNumber)

```
curl -s -X GET \
  "http://localhost:4000/channels/mychannel/blocks/1?peer=peer1" \
  -H "authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0OTQ4NjU1OTEsInVzZXJuYW1lIjoiSmltIiwib3JnTmFtZSI6Im9yZzEiLCJpYXQiOjE0OTQ4NjE5OTF9.yWaJhFDuTvMQRaZIqg20Is5t-JJ_1BP58yrNLOKxtNI" \
  -H "content-type: application/json"
```

### TransactionID로 트랜잭션 쿼리하기(Query Transaction by TransactionID) 

```
curl -s -X GET http://localhost:4000/channels/mychannel/transactions/TRX_ID?peer=peer1 \
  -H "authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0OTQ4NjU1OTEsInVzZXJuYW1lIjoiSmltIiwib3JnTmFtZSI6Im9yZzEiLCJpYXQiOjE0OTQ4NjE5OTF9.yWaJhFDuTvMQRaZIqg20Is5t-JJ_1BP58yrNLOKxtNI" \
  -H "content-type: application/json"
```
**참고**: 여기서 TRX_ID는 이전 호출 트랜잭션


### ChainInfo 쿼리하기(Query ChainInfo) 

```
curl -s -X GET \
  "http://localhost:4000/channels/mychannel?peer=peer1" \
  -H "authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0OTQ4NjU1OTEsInVzZXJuYW1lIjoiSmltIiwib3JnTmFtZSI6Im9yZzEiLCJpYXQiOjE0OTQ4NjE5OTF9.yWaJhFDuTvMQRaZIqg20Is5t-JJ_1BP58yrNLOKxtNI" \
  -H "content-type: application/json"
```

### 설치된 체인 코드 쿼리하기(Query Installed chaincodes) 

```
curl -s -X GET \
  "http://localhost:4000/chaincodes?peer=peer1&type=installed" \
  -H "authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0OTQ4NjU1OTEsInVzZXJuYW1lIjoiSmltIiwib3JnTmFtZSI6Im9yZzEiLCJpYXQiOjE0OTQ4NjE5OTF9.yWaJhFDuTvMQRaZIqg20Is5t-JJ_1BP58yrNLOKxtNI" \
  -H "content-type: application/json"
```

### 인스턴스화된 체인 코드 쿼리하기(Query Instantiated chaincodes) 

```
curl -s -X GET \
  "http://localhost:4000/chaincodes?peer=peer1&type=instantiated" \
  -H "authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0OTQ4NjU1OTEsInVzZXJuYW1lIjoiSmltIiwib3JnTmFtZSI6Im9yZzEiLCJpYXQiOjE0OTQ4NjE5OTF9.yWaJhFDuTvMQRaZIqg20Is5t-JJ_1BP58yrNLOKxtNI" \
  -H "content-type: application/json"
```

### 채널 쿼리하기(Query Channels) 

```
curl -s -X GET \
  "http://localhost:4000/channels?peer=peer1" \
  -H "authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0OTQ4NjU1OTEsInVzZXJuYW1lIjoiSmltIiwib3JnTmFtZSI6Im9yZzEiLCJpYXQiOjE0OTQ4NjE5OTF9.yWaJhFDuTvMQRaZIqg20Is5t-JJ_1BP58yrNLOKxtNI" \
  -H "content-type: application/json"
```

### 네트워크 구성 고려 사항(Network configuration considerations) 

network-config.json 파일을 직접 편집하거나 대체 대상 네트워크에 대한 추가 파일을 제공하여 구성 매개 변수를 변경할 수 있습니다. 응용 프로그램은 선택적 환경 변수 "TARGET_NETWORK"을 사용하여 사용할 구성 파일을 제어합니다. 예를 들어 Amazon Web Services EC2에 대상 네트워크를 배포 한 경우 "network-config-aws.json"파일을 추가하고 "TARGET_NETWORK"환경을 'aws'로 설정할 수 있습니다. 응용 프로그램은 "network-config-aws.json"파일 내에서 설정을 선택합니다.

#### IP 주소 ** 및 PORT 정보(IP Address** and PORT information) 

동료와 orderer의 IP 주소 및 PORT 정보를 하드 코딩하여 docker-compose yaml 파일을 사용자 정의하는 경우, 동일한 값을 network-config.json 파일에 추가해야합니다. 아래 표시된 경로는 도커 작성 yaml 파일과 일치하도록 조정해야합니다.

```
		"orderer": {
			"url": "grpcs://x.x.x.x:7050",
			"server-hostname": "orderer0",
			"tls_cacerts": "../artifacts/tls/orderer/ca-cert.pem"
		},
		"org1": {
			"ca": "http://x.x.x.x:7054",
			"peer1": {
				"requests": "grpcs://x.x.x.x:7051",
				"events": "grpcs://x.x.x.x:7053",
				...
			},
			"peer2": {
				"requests": "grpcs://x.x.x.x:7056",
				"events": "grpcs://x.x.x.x:7058",
				...
			}
		},
		"org2": {
			"ca": "http://x.x.x.x:8054",
			"peer1": {
				"requests": "grpcs://x.x.x.x:8051",
				"events": "grpcs://x.x.x.x:8053",
				...			},
			"peer2": {
				"requests": "grpcs://x.x.x.x:8056",
				"events": "grpcs://x.x.x.x:8058",
				...
			}
		}

```

#### IP 주소 검색(Discover IP Address) 

네트워크 엔티티 중 하나의 IP 주소를 검색하려면 다음 명령을 실행하십시오:

```
# this will return the IP Address for peer0
docker inspect peer0 | grep IPAddress
```

<a rel="license" href="http://creativecommons.org/licenses/by/4.0/"><img alt="Creative Commons License" style="border-width:0" src="https://i.creativecommons.org/l/by/4.0/88x31.png" /></a><br />This work is licensed under a <a rel="license" href="http://creativecommons.org/licenses/by/4.0/">Creative Commons Attribution 4.0 International License</a>.
