package main

import (
  "errors"
	"fmt"
	"strconv"
	"encoding/json"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type SimpleChaincode struct {
}

type Account struct {
	ID          string  `json:"ID"`
	CashBalance int     `json:"CashBalance"`
}

type Transaction struct {
	CUSIP       string   `json:"cusip"`
	FromUser    string   `json:"fromUser"`
	ToUser      string   `json:"toUser"`
	Quantity    int      `json:"quantity"`
}


// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	var Aval int
	var err error

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState("abc", []byte(strconv.Itoa(Aval)))				//making a test var "abc", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}

	// var empty []string
	// jsonAsBytes, _ := json.Marshal(empty)								//marshal an emtpy array of strings to clear the index
	// err = stub.PutState(marbleIndexStr, jsonAsBytes)
	// if err != nil {
	// 	return nil, err
	// }

	return nil, nil
}
// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}


// // ============================================================================================================================
// // Run - Our entry point for Invocations - [LEGACY] obc-peer 4/25/2016
// // ============================================================================================================================
// func (t *SimpleChaincode) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
// 	fmt.Println("run is running " + function)
// 	return t.Invoke(stub, function, args)
// }

// ============================================================================================================================
// Invoke - Our entry point for  TO REMOVE SHITTTTTT
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

  if function == "init" {													//initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	}  else if function == "write" {											//writes a value to the chaincode state
		return t.Write(stub, args)
	} else if function == "createAccount" {
    return t.createAccount(stub, args)
  } else if function == "set_user" {										//change owner of a marble
		res, err := t.set_user(stub, args)											//lets make sure all open trades are still valid
		return res, err
	}
	fmt.Println("invoke did not find func: " + function)					//error

	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" {													//read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) read(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil													//send it onward
}

func (t *SimpleChaincode) Write(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var name, value string // Entities
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
	}

	name = args[0]															//rename for funsies
	value = args[1]
	err = stub.PutState(name, []byte(value))								//write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (t *SimpleChaincode) createAccount(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
    // Obtain the username to associate with the account
    if len(args) != 1 {
        fmt.Println("Error obtaining username")
        return nil, errors.New("createAccount accepts a single username argument")
    }
    username := args[0]

    var account = Account{ID: username, CashBalance: 500}

    //build the  json string manually
	account := `{"ID": "` + username + `", "CashBalance": "` + 500 +  `"}`
	err = stub.PutState(username, []byte(account))
	if err != nil {
		return nil, err
	}


    fmt.Println("Attempting to get state of any existing account for " + account.ID)

    existingBytes, err := stub.GetState(username)

	  if err == nil {
        var accountData Account
        err = json.Unmarshal(existingBytes, &accountData)
        if err != nil {
            fmt.Println("Error unmarshalling account " + account.ID + "\n--->: " + err.Error())

            if strings.Contains(err.Error(), "unexpected end") {
                fmt.Println("No data means existing account found for " + account.ID + ", initializing account.")
                err = stub.PutState(username, []byte(account))

                if err == nil {
                    fmt.Println("created account" + account.ID)
                    return []byte(account), nil
                } else {
                    fmt.Println("failed to create initialize account for " + account.ID)
                    return nil, errors.New("failed to initialize an account for " + account.ID + " => " + err.Error())
                }
            } else {
                return nil, errors.New("Error unmarshalling existing account " + account.ID)
            }
        } else {
            fmt.Println("Account already exists for " + account.ID + " " + accountData.ID)
		        return nil, errors.New("Can't reinitialize existing user " + account.ID)
        }
    } else {
        fmt.Println("No existing account found for " + account.ID + ", initializing account.")
        err = stub.PutState(account.ID, []byte(account))

        if err == nil {
            fmt.Println("created account" + account.ID)
            return []byte(account), nil
        } else {
            fmt.Println("failed to create initialize account for " + account.ID)
            return nil, errors.New("failed to initialize an account for " + account.ID + " => " + err.Error())
        }
    }
}
// ============================================================================================================================
// Set Trade - create an open trade for a marble you want with marbles you have
// ============================================================================================================================
func (t *SimpleChaincode) set_user(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error
  var toRes Account
	//     0         1        2
	// "fromUser", "500", "toUser",
	if len(args) < 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	fromAccountAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get thing")
	}
  toAccountAsBytes, err := stub.GetState(args[2])
	if err != nil {
		return nil, errors.New("Failed to get thing")
	}


	fromRes := Account{}
	json.Unmarshal(fromAccountAsBytes, &fromRes)										//un stringify it aka JSON.parse()

  toRes = Account{}
	json.Unmarshal(toAccountAsBytes, &toRes)



	accountBalance := fromRes.CashBalance


  transferAmount, err := strconv.Atoi(args[1])
   if err != nil {
      // handle error
   }
  if(accountBalance < transferAmount) {
    fmt.Println("- Insufficient funds")
    return nil, nil
  }

  toRes.CashBalance = accountBalance + transferAmount
  fromRes.CashBalance = fromRes.CashBalance - transferAmount

	toJsonAsBytes, _ := json.Marshal(toRes)
	err = stub.PutState(args[2], toJsonAsBytes)								//rewrite the marble with id as key
	if err != nil {
		return nil, err
	}

  fromJsonAsBytes, _ := json.Marshal(fromRes)
	err = stub.PutState(args[0], fromJsonAsBytes)								//rewrite the marble with id as key
	if err != nil {
		return nil, err
	}

	fmt.Println("- end set trade")
	return nil, nil
}
