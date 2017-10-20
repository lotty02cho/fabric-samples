/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

//WARNING - this chaincode's ID is hard-coded in chaincode_example04 to illustrate one way of
//calling chaincode from a chaincode. If this example is modified, chaincode_example04.go has
//to be modified as well with the new ID of chaincode_example02.
//chaincode_example05 show's how chaincode ID can be passed in as a parameter instead of
//hard-coding.

//(경고 -이 chaincode의 ID는 chaincode_example04에 하드 코딩되어 있습니다.
// chaincode에서 chaincode를 호출합니다. 이 예제를 수정하면
// chaincode_example04.go를chaincode_example02의 새 ID로 수정해야합니다.
// chaincode_example05 show는 체인 코드 ID가 하드 코딩 대신 매개 변수로 전달되는 방법을 보여줍니다.)

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
// (SimpleChaincode 예제 간단한 체인 코드 구현)
type SimpleChaincode struct {
}

/////////////Init(초기화)/////////////
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("ex02 Init")
	_, args := stub.GetFunctionAndParameters()
	var A, B string    // Entities(개체)
	var Aval, Bval int // Asset holdings(보유 자산)
	var err error

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	// Initialize the chaincode(체인코드 초기화합니다.)
	A = args[0]
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	B = args[2]
	Bval, err = strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state to the ledger(원장에다가 state를 쓰십시오.)
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

/////////////Invoke(호출)/////////////
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("ex02 Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "invoke" {
		// Make payment of X units from A to B(X 값을 A에서 B로 지불하십시오.)
		return t.invoke(stub, args)
	} else if function == "delete" {
		// Deletes an entity from its state(현재의 상태에서 엔티티를 삭제합니다.)
		return t.delete(stub, args)
	} else if function == "query" {
		// the old "Query" is now implemtned in invoke(이전의 "Query"는 이제 invoke에서 구현됩니다.)
		return t.query(stub, args)
	} else if function == "reg" {
		return t.reg(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"invoke\" \"delete\" \"query\"")
}

/////////////Invoke(호출)/////////////
// Transaction makes payment of X units from A to B(거래는 A에서 B로 X 값을 지불합니다.)
func (t *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A, B string    // Entities(개체)
	var Aval, Bval int // Asset holdings(보유 자산)
	var X int          // Transaction value(거래 값)
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	A = args[0]
	B = args[1]

	// Get the state from the ledger(장부에서 상태를 가져옵니다.)
	// TODO: will be nice to have a GetAllState call to ledger(장부에 GetAllState 호출을 하는 것이 좋을 것입니다.)
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return shim.Error("Entity not found")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution(실행 수행)
	X, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	Aval = Aval - X
	Bval = Bval + X
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state back to the ledger(원장에다가 상태를 다시 쓰십시오.)
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

/////////////Delete(삭제)/////////////
// Deletes an entity from state(상태에서 엔티티를 삭제합니다.)
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	A := args[0]

	// Delete the key from the state in ledger(원장에서 키를 삭제합니다.)
	err := stub.DelState(A)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

/////////////Query(쿼리)/////////////
// query callback representing the query of a chaincode(체인 코드 쿼리를 나타내는 쿼리 콜백)
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string // Entities(객체)
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

	// Get the state from the ledger(장부에서 상태를 가져옵니다.)
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// register an entity from state(상태에서 엔티티 등록)
func (t *SimpleChaincode) reg(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string //Entities(객체)
	var Aval int //Asset holdings(보유 자산)
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	//Initialize the chaincode(체인코드를 초기화)
	//객체 값 넣어주고,
	A = args[0]
	// 보유자산 값 넣어주고,
	Aval, err = strconv.Atoi(args[1])
	if err != nil{
		return shim.Error("Expecting integer value for asset holding")
	}

	//문제가 없으면 새로운 값 출력
	fmt.Printf("New Value = %d\n", Aval)

	//Write the state to the ledger(블록체인에 객체와 객체가 가진 값을 넣어준다.)
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}