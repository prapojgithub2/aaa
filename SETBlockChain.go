package main

import (
	//"encoding/base64"
	//"encoding/binary"
	"encoding/json"
	"errors"
	"strconv"
	//"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/op/go-logging"
)

var myLogger = logging.MustGetLogger("setBlockChain")

var txHandler = NewTransactionHandler()
var actBalHandler = NewAccountBalanceHandler()
var actMonHandler = NewAccountMoneyHandler()
var secProHandler = NewSecurityProfileHandler()

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

	if accountid != txMsg.BuyerID {
		return nil, errors.New("Invalid buyerID")
	}

	price, err := strconv.ParseUint(txMsg.Price, 10, 64)
	if err != nil {
		myLogger.Errorf("system error %v", err)
		return nil, errors.New("Unable to parse Price" + txMsg.Price)
	}
	var noOfHolderAllowed uint64 = 5;
	myLogger.Debugf("+++++++++++++++++++++++++++++++++++ termsheet validation +++++++++++++++++++++++++++++++++")	
	ok , err := actBalHandler.validateOverTermSheetRules(stub,txMsg.SellerID,txMsg.BuyerID,txMsg.Symbol,txMsg.Volume,noOfHolderAllowed);	
	if ( ok ){
		myLogger.Debugf("+++++++++++++++++++++++++++++++++++ validate OK +++++++++++++++++++++++++++++++++")
		actMonHandler.transfer(stub, txMsg.BuyerID, txMsg.SellerID, price*txMsg.Volume)
		actBalHandler.transferAccountBalance(stub, txMsg.SellerID, txMsg.BuyerID, txMsg.Symbol, txMsg.Volume)
		return nil, txHandler.updateStatus(stub, txID, STATUS_CONFIRMED)
	} else {
		myLogger.Debugf("+++++++++++++++++++++++++++++++++++ validate FAILS +++++++++++++++++++++++++++++++++")
		return nil, errors.New("Not pass termsheet validation");
	}

}

func (t *SETBlockChainChaincode) findUnconfirmedTransaction(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debugf("+++++++++++++++++++++++++++++++++++ findUnconfirmedTransaction +++++++++++++++++++++++++++++++++")

	if len(args) != 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}

	accountid, err := t.getCertAttribute(stub)
	if err != nil {
		return nil, err
	}
	myLogger.Debugf("accountid [%v]", accountid)

	txMsgs, err := txHandler.findTransactionByAccountID(stub, accountid)
	if err != nil {
		return nil, err
	}

	var txMsgsUnconfirmed []TransactionMsg

	for _, txMsg := range txMsgs {
		if txMsg.BuyerID == accountid && txMsg.Status == STATUS_WAITING {
			txMsgsUnconfirmed = append(txMsgsUnconfirmed, txMsg)
		}
	}

	txMsgsJson, err := json.Marshal(txMsgsUnconfirmed)
	myLogger.Debugf("Response : %s", txMsgsJson)

	return txMsgsJson, nil
}

func (t *SETBlockChainChaincode) findConfirmedTransactionBySymbol(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debugf("+++++++++++++++++++++++++++++++++++ findConfirmedTransactionBySymbol +++++++++++++++++++++++++++++++++")

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	accountid, err := t.getCertAttribute(stub)
	if err != nil {
		return nil, err
	}
	myLogger.Debugf("accountid [%v]", accountid)

	var symbol string
	symbol = args[0]

	txMsgs, err := txHandler.findTransactionByAccountIDSymbol(stub, accountid, symbol)
	if err != nil {
		return nil, err
	}

	var txMsgsConfirmed []TransactionMsg

	for _, txMsg := range txMsgs {
		if txMsg.Status == STATUS_CONFIRMED {
			txMsgsConfirmed = append(txMsgsConfirmed, txMsg)
		}
	}

	txMsgsJson, err := json.Marshal(txMsgsConfirmed)
	myLogger.Debugf("Response : %s", txMsgsJson)

	return txMsgsJson, nil
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

	txMsgs, err := txHandler.findTransactionByAccountID(stub, accountid)
	if err != nil {
		return nil, err
	}

	for index, txMsg := range txMsgs {
		myLogger.Debugf("[%v]:[%v]", index, txMsg)
	}

	return txHandler.query(stub, accountid)
}

func (t *SETBlockChainChaincode) getBalance(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debugf("+++++++++++++++++++++++++++++++++++ getBalance +++++++++++++++++++++++++++++++++")

	if len(args) != 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}

	accountid, err := t.getCertAttribute(stub)
	if err != nil {
		return nil, err
	}
	myLogger.Debugf("accountid [%v]", accountid)

	return actBalHandler.query(stub, accountid)
}

func (t *SETBlockChainChaincode) getMoney(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debugf("+++++++++++++++++++++++++++++++++++getMoney+++++++++++++++++++++++++++++++++")

	accountid, err := t.getCertAttribute(stub)
	if err != nil {
		return nil, err
	}

	balance, err := actMonHandler.queryBalance(stub, accountid)
	if err != nil {
		return nil, err
	}

	var txMsgs []AccountMoneyMsg
	txMsg := AccountMoneyMsg{
		accountid, //AccountID
		balance,   //Amount
	}
	txMsgs = append(txMsgs, txMsg)

	txMsgsJSON, err := json.Marshal(txMsgs)
	myLogger.Debugf("Response : %s", txMsgsJSON)

	//return []byte(balanceStr), nil
	return txMsgsJSON, nil
}

func (t *SETBlockChainChaincode) getMaxNumberHolder(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debugf("+++++++++++++++++++++++++++++++++++getMaxNumberHolder+++++++++++++++++++++++++++++++++")
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}

	sysmbol := args[0]

	maxNumberHolder, err := secProHandler.getMaxNumberHolder(stub, sysmbol)
	if err != nil {
		return nil, err
	}

	var txMsgs []SecurityProfileMsg
	txMsg := SecurityProfileMsg{
		sysmbol,         //Symbol
		maxNumberHolder, //MaxNumberHolder
	}
	txMsgs = append(txMsgs, txMsg)

	txMsgsJSON, err := json.Marshal(txMsgs)
	myLogger.Debugf("Response : %s", txMsgsJSON)

	return txMsgsJSON, nil
}

func (t *SETBlockChainChaincode) Init(stub shim.ChaincodeStubInterface) ([]byte, error) {
	myLogger.Debugf("******************************** Init ****************************************")

	function, args := stub.GetFunctionAndParameters()

	myLogger.Infof("[SETBlockChainChaincode] Init[%v]", function)
	if len(args) != 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}
	/*test*/
	actBalHandler.createTable(stub)
	actMonHandler.createTable(stub)
	secProHandler.createTable(stub)
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
		*/
	} else if function == "confirmBuy" {
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
	} else if function == "getBalance" {
		return t.getBalance(stub, args)
	} else if function == "findUnconfirmedTransaction" {
		return t.findUnconfirmedTransaction(stub, args)
	} else if function == "findConfirmedTransactionBySymbol" {
		return t.findConfirmedTransactionBySymbol(stub, args)
	} else if function == "getMoney" {
		return t.getMoney(stub, args)
	} else if function == "getMaxNumberHolder" {
		return t.getMaxNumberHolder(stub, args)
	}

	return nil, errors.New("Received unknown function query invocation with function " + function)
}

func main() {

	//  primitives.SetSecurityLevel("SHA3", 256)
	err := shim.Start(new(SETBlockChainChaincode))
	if err != nil {
		myLogger.Debugf("Error starting SETBlockChainChaincode: %s", err)
	}

}
