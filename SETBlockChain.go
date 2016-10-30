package main

import (
  //"encoding/base64"
  //"encoding/binary"
  "errors"
  "strconv"
  //"strings"

  "github.com/hyperledger/fabric/core/chaincode/shim"
  "github.com/op/go-logging"
)

var myLogger = logging.MustGetLogger("setBlockChain")

var txHandler = NewTransactionHandler()

type SETBlockChainChaincode struct {
}

func (t *SETBlockChainChaincode) getCertAttribute(stub shim.ChaincodeStubInterface) (string, error) {
  accountid, err := stub.ReadCertAttribute("accountid")
  if err != nil {
    myLogger.Errorf("Cannot getCertAttribute : %v", err)
    return "", errors.New("Cannot getCertAttribute")
  }

  return string(accountid), nil
}

func (t *SETBlockChainChaincode) sell(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
  myLogger.Debugf("+++++++++++++++++++++++++++++++++++ sell +++++++++++++++++++++++++++++++++")

  if len(args) != 4 {
    return nil, errors.New("Incorrect number of arguments. Expecting 4")
  }

  accountid, err := t.getCertAttribute(stub)
  if err != nil {
    return nil, err
  }
  myLogger.Debugf("accountid [%v]", accountid)

  var symbol, buyerID, price string

  symbol = args[0]
  buyerID = args[1]
  price = args[2]
  volume, err := strconv.ParseUint(args[3], 10, 64)
  if err != nil {
    return nil, errors.New("Cannot parse volume")
  }

  // return nil, txHandler.insert(stub, "abc", "0001", "0002", []byte(strconv.Itoa(10)), 100, "WAITING")
  return nil, txHandler.insert(stub, symbol, buyerID, accountid, price, volume, STATUS_WAITING)
}

func (t *SETBlockChainChaincode) confirmBuy(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
  myLogger.Debugf("+++++++++++++++++++++++++++++++++++ confirmBuy +++++++++++++++++++++++++++++++++")

  if len(args) != 1 {
    return nil, errors.New("Incorrect number of arguments. Expecting 1")
  }

  accountid, err := t.getCertAttribute(stub)
  if err != nil {
    return nil, err
  }
  myLogger.Debugf("accountid [%v]", accountid)

  txID, err := strconv.ParseUint(args[0], 10, 64)
  if err != nil {
    return nil, errors.New("Cannot parse txID")
  }

  txMsg, err := txHandler.getTransaction(stub, txID)
  if txMsg == nil || err != nil {
    return nil, errors.New("Cannot find transaction")
  }

  myLogger.Debugf("BuyerID[%v]", txMsg.BuyerID)

  if(accountid != txMsg.BuyerID) {
    return nil, errors.New("Invalid buyerID")
  }

  return nil, txHandler.updateStatus(stub, txID, STATUS_CONFIRMED)
}

func (t *SETBlockChainChaincode) getTransaction(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
  myLogger.Debugf("+++++++++++++++++++++++++++++++++++ getTransaction +++++++++++++++++++++++++++++++++")

  if len(args) != 0 {
    return nil, errors.New("Incorrect number of arguments. Expecting 0")
  }

  accountid, err := t.getCertAttribute(stub)
  if err != nil {
    return nil, err
  }
  myLogger.Debugf("accountid [%v]", accountid)

  return txHandler.query(stub, accountid)
}

func (t *SETBlockChainChaincode) Init(stub shim.ChaincodeStubInterface) ([]byte, error) {
  myLogger.Debugf("******************************** Init ****************************************")

  function, args := stub.GetFunctionAndParameters()

  myLogger.Infof("[SETBlockChainChaincode] Init[%v]", function)
  if len(args) != 0 {
    return nil, errors.New("Incorrect number of arguments. Expecting 0")
  }

  return nil, txHandler.createTable(stub)
}

func (t *SETBlockChainChaincode) Invoke(stub shim.ChaincodeStubInterface) ([]byte, error) {
  myLogger.Debugf("******************************** Invoke ****************************************")

  function, args := stub.GetFunctionAndParameters()

  myLogger.Infof("[SETBlockChainChaincode] Invoke[%v]", function)

  //   Handle different functions
  if function == "sell" {
    return t.sell(stub, args)
  /*} else if function == "buy" {
    return t.buy(stub, args)
  } else if function == "confirmSell" {
    return t.confirmSell(stub, args)
  */} else if function == "confirmBuy" {
    return t.confirmBuy(stub, args)
  }

  return nil, errors.New("Received unknown function invocation")
}

func (t *SETBlockChainChaincode) Query(stub shim.ChaincodeStubInterface) ([]byte, error) {
  myLogger.Debugf("******************************** Query ****************************************")

  function, args := stub.GetFunctionAndParameters()

  myLogger.Infof("[SETBlockChainChaincode] Query[%v]", function)

  // Handle different functions
  if function == "getTransaction" {
    return t.getTransaction(stub, args)
  }/* else if function == "getBalance" {
    return t.getBalance(stub, args)
  }*/

  return nil, errors.New("Received unknown function query invocation with function " + function)
}

func main() {

  //  primitives.SetSecurityLevel("SHA3", 256)
  err := shim.Start(new(SETBlockChainChaincode))
  if err != nil {
    myLogger.Debugf("Error starting SETBlockChainChaincode: %s", err)
  }

}
