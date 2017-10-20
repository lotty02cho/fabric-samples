/*
 * Demonstrates how to handle data in an application with a high transaction volume where the transactions
 * all attempt to change the same key-value pair in the ledger. Such an application will have trouble
 * as multiple transactions may read a value at a certain version, which will then be invalid when the first
 * transaction updates the value to a new version, thus rejecting all other transactions until they're
 * re-executed.
 * Rather than relying on serialization of the transactions, which is slow, this application initializes
 * a value and then accepts deltas of that value which are added as rows to the ledger. The actual value
 * is then an aggregate of the initial value combined with all of the deltas. Additionally, a pruning
 * function is provided which aggregates and deletes the deltas to update the initial value. This should
 * be done during a maintenance window or when there is a lowered transaction volume, to avoid the proliferation
 * of millions of rows of data.
 * (트랜잭션이 모두 원장의 동일한 키 - 값 쌍을 변경하려고하는 높은 트랜잭션 볼륨을 가진 응용 프로그램에서 데이터를 처리하는 방법을 보여줍니다.
 * 이러한 응용 프로그램은 여러 버전의 트랜잭션이 특정 버전의 값을 읽을 수 있으므로 문제가됩니다. 첫 번째 트랜잭션이 값을 새 버전으로 업데이트
 * 할 때 유효하지 않으므로 다시 실행될 때까지 다른 모든 트랜잭션은 거부됩니다)
 * 
 * 느린 트랜잭션의 직렬화에 의존하는 대신이 응용 프로그램은 값을 초기화 한 다음 장부에 행으로 추가되는 해당 값의 델타를 받습니다.
 * 실제 값은 모든 델타와 결합 된 초기 값의 집합입니다. 또한 델타를 집계하고 삭제하여 초기 값을 갱신하는 prune 기능이 제공됩니다.
 * 이는 수백만 행의 데이터가 확산되는 것을 방지하기 위해 유지 관리 기간 또는 거래량이 감소한 경우에 수행해야합니다.
 *
 * @author	Alexandre Pauwels for IBM
 * @created	17 Aug 2017
 */

package main

/* Imports
 * 4 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
 * 2 specific Hyperledger Fabric specific libraries for Smart Contracts
 * (포맷팅, 바이트 처리, JSON 읽기 및 쓰기, 문자열 조작을위한 * 4 유틸리티 라이브러리
 * 스마트 계약을위한 2 가지 고유 Hyperbelger Fabric 라이브러리)
 */
import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

//SmartContract is the data structure which represents this contract and on which  various contract lifecycle functions are attached
//(SmartContract는이 계약을 나타내며 다양한 계약 라이프 사이클 기능이 첨부 된 데이터 구조입니다.)
type SmartContract struct {
}

// Define Status codes for the response
// (응답의 상태 코드 정의)
const (
	OK    = 200
	ERROR = 500
)

// Init is called when the smart contract is instantiated
// (스마트 계약이 인스턴스화 될 때 Init가 호출됩니다.)
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

// Invoke routes invocations to the appropriate function in chaincode
// (chaincode의 적절한 함수에 라우트 호출을 호출합니다.)
// Current supported invocations are:(현재 지원되는 호출은 다음과 같습니다.)
//	- update, adds a delta to an aggregate variable in the ledger, all variables are assumed to start at 0
//    (업데이트, 원장의 집계 변수에 델타를 추가합니다. 모든 변수는 0에서 시작한다고 가정합니다.)
//	- get, retrieves the aggregate value of a variable in the ledger
//    (get, ledger에서 변수의 집계 값을 검색합니다.)
//	- pruneFast, deletes all rows associated with the variable and replaces them with a single row containing the aggregate value
//    (pruneFast는 변수와 연관된 모든 행을 삭제하고 집계 값을 포함하는 단일 행으로 대체합니다.)
//	- pruneSafe, same as pruneFast except it pre-computed the value and backs it up before performing any destructive operations
//    (pruneSafe는 pruneFast와 동일하지만 값을 사전 계산하고 파괴적인 작업을 수행하기 전에 백업합니다.)
//	- delete, removes all rows associated with the variable
//    (변수와 관련된 모든 행을 삭제하고 삭제합니다.)
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	// Retrieve the requested Smart Contract function and arguments
	// (요청된 스마트 계약 함수 및 인수 검색)
	function, args := APIstub.GetFunctionAndParameters()

	// Route to the appropriate handler function to interact with the ledger appropriately
	// (원장과 적절하게 상호 작용하기 위해 적절한 핸들러 함수로 전달합니다.)
	if function == "update" {
		return s.update(APIstub, args)
	} else if function == "get" {
		return s.get(APIstub, args)
	} else if function == "prunefast" {
		return s.pruneFast(APIstub, args)
	} else if function == "prunesafe" {
		return s.pruneSafe(APIstub, args)
	} else if function == "delete" {
		return s.delete(APIstub, args)
	} else if function == "putstandard" {
		return s.putStandard(APIstub, args)
	} else if function == "getstandard" {
		return s.getStandard(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

/**
 * Updates the ledger to include a new delta for a particular variable. If this is the first time
 * this variable is being added to the ledger, then its initial value is assumed to be 0. The arguments
 * to give in the args array are as follows:
 * (원장이 특정 변수에 대한 새 델타를 포함하도록 업데이트합니다. 이번이 처음이라면 이 변수가 원장에 추가되면 초기 값은 0으로 간주됩니다.
 *  args 배열에 넣는 것은 다음과 같습니다 :)
 *	- args[0] -> name of the variable(변수의 이름)
 *	- args[1] -> new delta (float)(새로운 델타(float))
 *	- args[2] -> operation (currently supported are addition "+" and subtraction "-") (작동(현재 지원되는 추가 "+"및 빼기 "-"))
 *
 * @param APIstub The chaincode shim(APIstub 체인 코드 심)
 * @param args The arguments array for the update invocation(param args update 호출의 arguments 배열)
 *
 * @return A response structure indicating success or failure with a message
 * (메세지로 성공 또는 실패를 나타내는 A 응답 구조를 리턴합니다.)
 */
func (s *SmartContract) update(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// Check we have a valid number of args(유효한 수의 arg가 있는지 확인하십시오.)
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments, expecting 3")
	}

	// Extract the args(args를 추출하십시오.)
	name := args[0]
	op := args[2]
	_, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return shim.Error("Provided value was not a number")
	}

	// Make sure a valid operator is provided(유효한 연산자가 제공되는지 확인하십시오.)
	if op != "+" && op != "-" {
		return shim.Error(fmt.Sprintf("Operator %s is unrecognized", op))
	}

	// Retrieve info needed for the update procedure(업데이트 절차에 필요한 정보 검색)
	txid := APIstub.GetTxID()
	compositeIndexName := "varName~op~value~txID"

	// Create the composite key that will allow us to query for all deltas on a particular variable
	// (특정 변수의 모든 델타를 쿼리 할 수있는 복합 키를 만듭니다.)
	compositeKey, compositeErr := APIstub.CreateCompositeKey(compositeIndexName, []string{name, op, args[1], txid})
	if compositeErr != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s: %s", name, compositeErr.Error()))
	}

	// Save the composite key index(복합 키 인덱스를 저장합니다)
	compositePutErr := APIstub.PutState(compositeKey, []byte{0x00})
	if compositePutErr != nil {
		return shim.Error(fmt.Sprintf("Could not put operation for %s in the ledger: %s", name, compositePutErr.Error()))
	}

	return shim.Success([]byte(fmt.Sprintf("Successfully added %s%s to %s", op, args[1], name)))
}

/**
 * Retrieves the aggregate value of a variable in the ledger. Gets all delta rows for the variable
 * and computes the final value from all deltas. The args array for the invocation must contain the
 * following argument:
 *	- args[0] -> The name of the variable to get the value of
 *
 * @param APIstub The chaincode shim(APIstub 체인 코드 심)
 * @param args The arguments array for the get invocation(param args get 호출의 arguments 배열)
 *
 * @return A response structure indicating success or failure with a message
 * (메세지로 성공 또는 실패를 나타내는 A 응답 구조를 리턴합니다.)
 */
func (s *SmartContract) get(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// Check we have a valid number of args(유효한 수의 arg가 있는지 확인하십시오.)
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, expecting 1")
	}

	name := args[0]
	// Get all deltas for the variable(변수에 대한 모든 델타 가져 옵니다.)
	deltaResultsIterator, deltaErr := APIstub.GetStateByPartialCompositeKey("varName~op~value~txID", []string{name})
	if deltaErr != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve value for %s: %s", name, deltaErr.Error()))
	}
	defer deltaResultsIterator.Close()

	// Check the variable existed(존재하는 변수를 확인합니다.)
	if !deltaResultsIterator.HasNext() {
		return shim.Error(fmt.Sprintf("No variable by the name %s exists", name))
	}

	// Iterate through result set and compute final value(결과 집합을 반복하고 최종 값을 계산합니다.)
	var finalVal float64
	var i int
	for i = 0; deltaResultsIterator.HasNext(); i++ {
		// Get the next row(다음 행을 가져옵니다.)
		responseRange, nextErr := deltaResultsIterator.Next()
		if nextErr != nil {
			return shim.Error(nextErr.Error())
		}

		// Split the composite key into its component parts
		// (합성 키를 구성 요소 파트로 분할합니다.)
		_, keyParts, splitKeyErr := APIstub.SplitCompositeKey(responseRange.Key)
		if splitKeyErr != nil {
			return shim.Error(splitKeyErr.Error())
		}

		// Retrieve the delta value and operation(델타 값 및 연산을 검색합니다.)
		operation := keyParts[1]
		valueStr := keyParts[2]

		// Convert the value string and perform the operation(값 문자열을 변환하고 작업을 수행합니다.)
		value, convErr := strconv.ParseFloat(valueStr, 64)
		if convErr != nil {
			return shim.Error(convErr.Error())
		}

		switch operation {
		case "+":
			finalVal += value
		case "-":
			finalVal -= value
		default:
			return shim.Error(fmt.Sprintf("Unrecognized operation %s", operation))
		}
	}

	return shim.Success([]byte(strconv.FormatFloat(finalVal, 'f', -1, 64)))
}

/**
 * Prunes a variable by deleting all of its delta rows while computing the final value. Once all rows
 * have been processed and deleted, a single new row is added which defines a delta containing the final
 * computed value of the variable. This function is NOT safe as any failures or errors during pruning
 * will result in an undefined final value for the variable and loss of data. Use pruneSafe if data
 * integrity is important. The args array contains the following argument:
 *	- args[0] -> The name of the variable to prune
 * 최종 값을 계산하는 동안 모든 델타 행을 삭제하여 변수를 잘라냅니다. 일단 모든 행이 처리되고 삭제되면 변수의 최종 계산 된 값을
 * 포함하는 델타를 정의하는 하나의 새로운 행이 추가됩니다. prune 중에 오류나 오류가 발생하면 변수의 최종 값이 정의되지 않고 데이터가
 * 유실되므로이 함수는 안전하지 않습니다. 데이터 무결성이 중요한 경우 pruneSafe를 사용하십시오. args 배열에는 다음과 같은 인수가 있습니다.
 *  - args [0] -> 제거 할 변수의 이름
 *
 * @param APIstub The chaincode shim(APIstub 체인 코드 심)
 * @param args The arguments array for the get invocation(param args pruneFast 호출을 위한 args 배열)
 *
 * @return A response structure indicating success or failure with a message
 * (메세지로 성공 또는 실패를 나타내는 A 응답 구조를 리턴합니다.)
 */
func (s *SmartContract) pruneFast(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// Check we have a valid number of ars(유효한 수의 ar을 가지고 있는지 확인합니다.)
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, expecting 1")
	}

	// Retrieve the name of the variable to prune(Prune할 변수 이름을 검색합니다.)
	name := args[0]

	// Get all delta rows for the variable(변수에 대한 모든 델타 행을 얻습니다.)
	deltaResultsIterator, deltaErr := APIstub.GetStateByPartialCompositeKey("varName~op~value~txID", []string{name})
	if deltaErr != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve value for %s: %s", name, deltaErr.Error()))
	}
	defer deltaResultsIterator.Close()

	// Check the variable existed(변수가 존재하는지 확인하십시오.)
	if !deltaResultsIterator.HasNext() {
		return shim.Error(fmt.Sprintf("No variable by the name %s exists", name))
	}

	// Iterate through result set computing final value while iterating and deleting each key
	// (각 키를 반복 및 삭제하면서 결과 집합을 반복하여 최종 값 계산합니다.)
	var finalVal float64
	var i int
	for i = 0; deltaResultsIterator.HasNext(); i++ {
		// Get the next row(다음 행을 얻어옵니다.)
		responseRange, nextErr := deltaResultsIterator.Next()
		if nextErr != nil {
			return shim.Error(nextErr.Error())
		}

		// Split the key into its composite parts(키를 복합체 부분으로 분할합니다.)
		_, keyParts, splitKeyErr := APIstub.SplitCompositeKey(responseRange.Key)
		if splitKeyErr != nil {
			return shim.Error(splitKeyErr.Error())
		}

		// Retrieve the operation and value(조작 및 값을 검색하십시오.)
		operation := keyParts[1]
		valueStr := keyParts[2]

		// Convert the value to a float(값을 float로 변환합니다.)
		value, convErr := strconv.ParseFloat(valueStr, 64)
		if convErr != nil {
			return shim.Error(convErr.Error())
		}

		// Delete the row from the ledger(원장에서 행을 삭제합니다.)
		deltaRowDelErr := APIstub.DelState(responseRange.Key)
		if deltaRowDelErr != nil {
			return shim.Error(fmt.Sprintf("Could not delete delta row: %s", deltaRowDelErr.Error()))
		}

		// Add the value of the deleted row to the final aggregate
		// (삭제 된 행의 값을 최종 집계에 추가합니다.)
		switch operation {
		case "+":
			finalVal += value
		case "-":
			finalVal -= value
		default:
			return shim.Error(fmt.Sprintf("Unrecognized operation %s", operation))
		}
	}

	// Update the ledger with the final value and return(원장을 최종 가치로 갱신하고 반환하십시오.)
	updateResp := s.update(APIstub, []string{name, strconv.FormatFloat(finalVal, 'f', -1, 64), "+"})
	if updateResp.Status == OK {
		return shim.Success([]byte(fmt.Sprintf("Successfully pruned variable %s, final value is %f, %d rows pruned", args[0], finalVal, i)))
	}

	return shim.Error(fmt.Sprintf("Failed to prune variable: all rows deleted but could not update value to %f, variable no longer exists in ledger", finalVal))
}

/**
 * This function performs the same function as pruneFast except it provides data backups in case the
 * prune fails. The final aggregate value is computed before any deletion occurs and is backed up
 * to a new row. This back-up row is deleted only after the new aggregate delta has been successfully
 * written to the ledger. The args array contains the following argument:
 *	args[0] -> The name of the variable to prune
 * 이 함수는 prune이 실패 할 경우 데이터 백업을 제공한다는 점을 제외하고는 pruneFast와 동일한 기능을 수행합니다. 
 * 최종 집계 값은 삭제가 발생하기 전에 계산되어 새 행에 백업됩니다. 이 백업 행은 새 집계 델타가 원장에 성공적으로 기록된 후에 만 삭제됩니다.
 * args 배열에는 다음과 같은 인수가 있습니다.
 *  args [0] -> 제거 할 변수의 이름
 *
 * @param APIstub The chaincode shim(APIstub 체인 코드 심)
 * @param args The arguments array for the get invocation(param args pruneSafe 호출을 위한 args 배열)
 *
 * @result A response structure indicating success or failure with a message
 * (메세지로 성공 또는 실패를 나타내는 A 응답 구조를 리턴합니다.)
 */
func (s *SmartContract) pruneSafe(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// Verify there are a correct number of arguments(올바른 수의 인수가 있는지 확인하십시오.)
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, expecting 1 (the name of the variable to prune)")
	}

	// Get the var name(var 이름 가져 오기)
	name := args[0]

	// Get the var's value and process it(var의 값을 가져 와서 처리하십시오.)
	getResp := s.get(APIstub, args)
	if getResp.Status == ERROR {
		return shim.Error(fmt.Sprintf("Could not retrieve the value of %s before pruning, pruning aborted: %s", name, getResp.Message))
	}

	valueStr := string(getResp.Payload)
	val, convErr := strconv.ParseFloat(valueStr, 64)
	if convErr != nil {
		return shim.Error(fmt.Sprintf("Could not convert the value of %s to a number before pruning, pruning aborted: %s", name, convErr.Error()))
	}

	// Store the var's value temporarily(var의 값을 임시로 저장하십시오.)
	backupPutErr := APIstub.PutState(fmt.Sprintf("%s_PRUNE_BACKUP", name), []byte(valueStr))
	if backupPutErr != nil {
		return shim.Error(fmt.Sprintf("Could not backup the value of %s before pruning, pruning aborted: %s", name, backupPutErr.Error()))
	}

	// Get all deltas for the variable(변수에 대한 모든 델타를 가져옵니다.)
	deltaResultsIterator, deltaErr := APIstub.GetStateByPartialCompositeKey("varName~op~value~txID", []string{name})
	if deltaErr != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve value for %s: %s", name, deltaErr.Error()))
	}
	defer deltaResultsIterator.Close()

	// Delete each row(각 행을 삭제합니다.)
	var i int
	for i = 0; deltaResultsIterator.HasNext(); i++ {
		responseRange, nextErr := deltaResultsIterator.Next()
		if nextErr != nil {
			return shim.Error(fmt.Sprintf("Could not retrieve next row for pruning: %s", nextErr.Error()))
		}

		deltaRowDelErr := APIstub.DelState(responseRange.Key)
		if deltaRowDelErr != nil {
			return shim.Error(fmt.Sprintf("Could not delete delta row: %s", deltaRowDelErr.Error()))
		}
	}

	// Insert new row for the final value(최종 값에 대한 새 행을 삽입합니다.)
	updateResp := s.update(APIstub, []string{name, valueStr, "+"})
	if updateResp.Status == ERROR {
		return shim.Error(fmt.Sprintf("Could not insert the final value of the variable after pruning, variable backup is stored in %s_PRUNE_BACKUP: %s", name, updateResp.Message))
	}

	// Delete the backup value(백업 값을 삭제합니다.)
	delErr := APIstub.DelState(fmt.Sprintf("%s_PRUNE_BACKUP", name))
	if delErr != nil {
		return shim.Error(fmt.Sprintf("Could not delete backup value %s_PRUNE_BACKUP, this does not affect the ledger but should be removed manually", name))
	}

	return shim.Success([]byte(fmt.Sprintf("Successfully pruned variable %s, final value is %f, %d rows pruned", name, val, i)))
}

/**
 * Deletes all rows associated with an aggregate variable from the ledger. The args array
 * contains the following argument:
 *	- args[0] -> The name of the variable to delete
 * 원장의 집계 변수와 관련된 모든 행을 삭제합니다. args 배열에는 다음 인수가 들어 있습니다.
 *  - args [0] -> 삭제할 변수의 이름
 *
 * @param APIstub The chaincode shim(APIstub 체인 코드 심)
 * @param args The arguments array for the delete invocation(param args delete 호출을 위한 args 배열)
 *
 * @return A response structure indicating success or failure with a message
 * (메세지로 성공 또는 실패를 나타내는 A 응답 구조를 리턴합니다.)
 */
func (s *SmartContract) delete(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// Check there are a correct number of arguments(올바른 수의 인수가 있는지 확인합니다.)
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, expecting 1")
	}

	// Retrieve the variable name(변수 이름 검색)
	name := args[0]

	// Delete all delta rows(모든 델타 행 삭제)
	deltaResultsIterator, deltaErr := APIstub.GetStateByPartialCompositeKey("varName~op~value~txID", []string{name})
	if deltaErr != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve delta rows for %s: %s", name, deltaErr.Error()))
	}
	defer deltaResultsIterator.Close()

	// Ensure the variable exists(변수가 있는지 확인합니다.)
	if !deltaResultsIterator.HasNext() {
		return shim.Error(fmt.Sprintf("No variable by the name %s exists", name))
	}

	// Iterate through result set and delete all indices(결과 집합을 반복하고 모든 인덱스를 삭제합니다.)
	var i int
	for i = 0; deltaResultsIterator.HasNext(); i++ {
		responseRange, nextErr := deltaResultsIterator.Next()
		if nextErr != nil {
			return shim.Error(fmt.Sprintf("Could not retrieve next delta row: %s", nextErr.Error()))
		}

		deltaRowDelErr := APIstub.DelState(responseRange.Key)
		if deltaRowDelErr != nil {
			return shim.Error(fmt.Sprintf("Could not delete delta row: %s", deltaRowDelErr.Error()))
		}
	}

	return shim.Success([]byte(fmt.Sprintf("Deleted %s, %d rows removed", name, i)))
}

/**
 * Converts a float64 to a byte array(float64를 바이트 배열로 변환합니다.)
 *
 * @param f The float64 to convert(param f, 변환할 float64)
 *
 * @return The byte array representation(반환 값, 바이트 배열 표현)
 */
func f2barr(f float64) []byte {
	str := strconv.FormatFloat(f, 'f', -1, 64)

	return []byte(str)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
// (주 기능은 단위 테스트 모드에서만 관련이 있습니다. 완전성을 위해 여기에 포함되었습니다.)
func main() {

	// Create a new Smart Contract(새로운 스마트 계약 만듭니다.)
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}

/**
 * All functions below this are for testing traditional editing of a single row(이 아래의 모든 기능은 단일 행의 기존 편집을 테스트하기위한 것입니다.)
 */
func (s *SmartContract) putStandard(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	name := args[0]
	valStr := args[1]

	_, getErr := APIstub.GetState(name)
	if getErr != nil {
		return shim.Error(fmt.Sprintf("Failed to retrieve the statr of %s: %s", name, getErr.Error()))
	}

	putErr := APIstub.PutState(name, []byte(valStr))
	if putErr != nil {
		return shim.Error(fmt.Sprintf("Failed to put state: %s", putErr.Error()))
	}

	return shim.Success(nil)
}

func (s *SmartContract) getStandard(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	name := args[0]

	val, getErr := APIstub.GetState(name)
	if getErr != nil {
		return shim.Error(fmt.Sprintf("Failed to get state: %s", getErr.Error()))
	}

	return shim.Success(val)
}
