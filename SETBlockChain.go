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

const (
	ROLE_ISSUER = "issuer"
	ROLE_TRADER = "trader"
	ROLE_BOT    = "bot"
	ROLE_TSD    = "tsd"
)

type SETBlockChainChaincode struct {
}

func (t *SETBlockChainChaincode) stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (t *SETBlockChainChaincode) getAccountid(stub shim.ChaincodeStubInterface) (string, error) {
	accountid, err := stub.ReadCertAttribute("accountid")
	if err != nil {
		myLogger.Errorf("Cannot getAccountid : %v", err)
		return "", errors.New("Cannot getAccountid")
	}

	return string(accountid), nil
}

func (t *SETBlockChainChaincode) getRole(stub shim.ChaincodeStubInterface) (string, error) {
	role, err := stub.ReadCertAttribute("role")
	if err != nil {
		myLogger.Errorf("Cannot getRole : %v", err)
		return "", errors.New("Cannot getRole")
	}

	return string(role), nil
}

func (t *SETBlockChainChaincode) sell(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debugf("+++++++++++++++++++++++++++++++++++ sell +++++++++++++++++++++++++++++++++")

	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	accountid, err := t.getAccountid(stub)
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

	accountid, err := t.getAccountid(stub)
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
	//var noOfHolderAllowed uint64 = 5;
	noOfHolderAllowed, err := secProHandler.getMaxNumberHolder(stub, txMsg.Symbol)
	if err != nil {
		return nil, err
	}
	myLogger.Infof("+++++++++++++++++++++++++++++++++++ termsheet validation +++++++++++++++++++++++++++++++++")
	myLogger.Debugf("noOfHolderAllowed [%v]", noOfHolderAllowed)
	ok, err := actBalHandler.validateOverTermSheetRules(stub, txMsg.SellerID, txMsg.BuyerID, txMsg.Symbol, txMsg.Volume, noOfHolderAllowed)
	if ok {
		myLogger.Infof("+++++++++++++++++++++++++++++++++++ validate OK +++++++++++++++++++++++++++++++++")
		actMonHandler.transfer(stub, txMsg.BuyerID, txMsg.SellerID, price*txMsg.Volume)
		actBalHandler.transferAccountBalance(stub, txMsg.SellerID, txMsg.BuyerID, txMsg.Symbol, txMsg.Volume)
		return nil, txHandler.updateStatus(stub, txID, STATUS_CONFIRMED)
	} else {
		myLogger.Infof("+++++++++++++++++++++++++++++++++++ validate FAILS +++++++++++++++++++++++++++++++++")
		return nil, errors.New("Not pass termsheet validation")
	}

}

func (t *SETBlockChainChaincode) cancel(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debugf("+++++++++++++++++++++++++++++++++++ cancel +++++++++++++++++++++++++++++++++")

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	accountid, err := t.getAccountid(stub)
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

	myLogger.Debugf("Status[%v]", txMsg.Status)

	if STATUS_WAITING != txMsg.Status {
		return nil, errors.New("Invalid Status")
	}

	myLogger.Debugf("BuyerID[%v]", txMsg.BuyerID)
	myLogger.Debugf("SellerID[%v]", txMsg.SellerID)

	if accountid == txMsg.BuyerID {
		return nil, txHandler.updateStatus(stub, txID, STATUS_CANCEL_BUYER)
	}

	if accountid == txMsg.SellerID {
		return nil, txHandler.updateStatus(stub, txID, STATUS_CANCEL_SELLER)
	}

	return nil, errors.New("Invalid buyerID or SellerID")
}

func (t *SETBlockChainChaincode) issueStock(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debugf("+++++++++++++++++++++++++++++++++++ issueStock +++++++++++++++++++++++++++++++++")

	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}

	accountid := args[0]
	symbol := args[1]
	volume, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		return nil, errors.New("Cannot parse volume")
	}

	return nil, actBalHandler.issueStock(stub, accountid, symbol, volume)
}

func (t *SETBlockChainChaincode) findUnconfirmedTransaction(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debugf("+++++++++++++++++++++++++++++++++++ findUnconfirmedTransaction +++++++++++++++++++++++++++++++++")

	if len(args) != 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}

	accountid, err := t.getAccountid(stub)
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
		if txMsg.Status == STATUS_WAITING {
			txMsgsUnconfirmed = append(txMsgsUnconfirmed, txMsg)
		}
	}

	txMsgsJson, err := json.Marshal(txMsgsUnconfirmed)
	myLogger.Debugf("Response : %s", txMsgsJson)

	return txMsgsJson, nil
}

func (t *SETBlockChainChaincode) findCompletedTransaction(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debugf("+++++++++++++++++++++++++++++++++++ findCompletedTransaction +++++++++++++++++++++++++++++++++")

	if len(args) != 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}

	accountid, err := t.getAccountid(stub)
	if err != nil {
		return nil, err
	}
	myLogger.Debugf("accountid [%v]", accountid)

	txMsgs, err := txHandler.findTransactionByAccountID(stub, accountid)
	if err != nil {
		return nil, err
	}

	var txMsgsCompleted []TransactionMsg

	for _, txMsg := range txMsgs {
		if txMsg.Status == STATUS_CONFIRMED || txMsg.Status == STATUS_CANCEL_BUYER || txMsg.Status == STATUS_CANCEL_SELLER {
			txMsgsCompleted = append(txMsgsCompleted, txMsg)
		}
	}

	txMsgsJson, err := json.Marshal(txMsgsCompleted)
	myLogger.Debugf("Response : %s", txMsgsJson)

	return txMsgsJson, nil
}

func (t *SETBlockChainChaincode) findConfirmedTransactionBySymbol(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debugf("+++++++++++++++++++++++++++++++++++ findConfirmedTransactionBySymbol +++++++++++++++++++++++++++++++++")

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	accountid, err := t.getAccountid(stub)
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

	accountid, err := t.getAccountid(stub)
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

	accountid, err := t.getAccountid(stub)
	if err != nil {
		return nil, err
	}
	myLogger.Debugf("accountid [%v]", accountid)

	return actBalHandler.query(stub, accountid)
}

func (t *SETBlockChainChaincode) getHolders(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debugf("+++++++++++++++++++++++++++++++++++ getHolders +++++++++++++++++++++++++++++++++")

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// accountid, err := t.getAccountid(stub)
	// if err != nil {
	// 	return nil, err
	// }
	// myLogger.Debugf("accountid [%v]", accountid)

	return actBalHandler.listHolderBySymbol(stub, args[0])
}

func (t *SETBlockChainChaincode) getMoney(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debugf("+++++++++++++++++++++++++++++++++++getMoney+++++++++++++++++++++++++++++++++")

	accountid, err := t.getAccountid(stub)
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

func (t *SETBlockChainChaincode) addMoney(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debugf("+++++++++++++++++++++++++++++++++++addMoney+++++++++++++++++++++++++++++++++")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	accountid := args[0]
	amount, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return nil, errors.New("Cannot parse volume")
	}

	return nil, actMonHandler.addMoney(stub, accountid, amount)
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

func (t *SETBlockChainChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	myLogger.Debugf("******************************** Init ****************************************")

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

func (t *SETBlockChainChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	myLogger.Debugf("******************************** Invoke ****************************************")

	myLogger.Infof("[SETBlockChainChaincode] Invoke[%v]", function)

	role, err := t.getRole(stub)
	if err != nil {
		return nil, err
	}

	//   Handle different functions
	if function == "sell" {
		if !t.stringInSlice(role, []string{ROLE_TRADER, ROLE_ISSUER}) {
			return nil, errors.New("Invalid role")
		}
		return t.sell(stub, args)
		/*} else if function == "buy" {
		    return t.buy(stub, args)
		  } else if function == "confirmSell" {
		    return t.confirmSell(stub, args)
		*/
	} else if function == "confirmBuy" {
		if !t.stringInSlice(role, []string{ROLE_TRADER, ROLE_ISSUER}) {
			return nil, errors.New("Invalid role")
		}
		return t.confirmBuy(stub, args)
	} else if function == "cancel" {
		if !t.stringInSlice(role, []string{ROLE_TRADER, ROLE_ISSUER}) {
			return nil, errors.New("Invalid role")
		}
		return t.cancel(stub, args)
	} else if function == "issueStock" {
		if !t.stringInSlice(role, []string{ROLE_TSD}) {
			return nil, errors.New("Invalid role")
		}
		return t.issueStock(stub, args)
	} else if function == "addMoney" {
		if !t.stringInSlice(role, []string{ROLE_BOT}) {
			return nil, errors.New("Invalid role")
		}
		return t.addMoney(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

type CallbackFunc func(stub shim.ChaincodeStubInterface, args []string) ([]byte, error)

func (t *SETBlockChainChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	myLogger.Debugf("******************************** Query ****************************************")

	myLogger.Infof("[SETBlockChainChaincode] Query[%v]", function)

	// Handle different functions
	// functionMap := map[string]CallbackFunc{
	// 	"getTransaction" : 	t.getTransaction,
	// 	"getBalance": t.getBalance,
	// 	"findUnconfirmedTransaction": t.findUnconfirmedTransaction,
	// 	"findCompletedTransaction": t.findCompletedTransaction,
	// 	"findConfirmedTransactionBySymbol": t.findConfirmedTransactionBySymbol,
	// 	"getMoney": t.getMoney,
	// 	"getMaxNumberHolder": t.getMaxNumberHolder,
	// }

	// m := functionMap[function]
	// m(stub, args)
	// m(func(stub shim.ChaincodeStubInterface,args []string))(stub, args)

	role, err := t.getRole(stub)
	if err != nil {
		return nil, err
	}

	if function == "getTransaction" {
		if !t.stringInSlice(role, []string{ROLE_TRADER, ROLE_ISSUER}) {
			return nil, errors.New("Invalid role")
		}
		return t.getTransaction(stub, args)
	} else if function == "getBalance" {
		if !t.stringInSlice(role, []string{ROLE_TRADER, ROLE_ISSUER}) {
			return nil, errors.New("Invalid role")
		}
		return t.getBalance(stub, args)
	} else if function == "findUnconfirmedTransaction" {
		if !t.stringInSlice(role, []string{ROLE_TRADER, ROLE_ISSUER}) {
			return nil, errors.New("Invalid role")
		}
		return t.findUnconfirmedTransaction(stub, args)
	} else if function == "findCompletedTransaction" {
		if !t.stringInSlice(role, []string{ROLE_TRADER, ROLE_ISSUER}) {
			return nil, errors.New("Invalid role")
		}
		return t.findCompletedTransaction(stub, args)
	} else if function == "findConfirmedTransactionBySymbol" {
		if !t.stringInSlice(role, []string{ROLE_TRADER, ROLE_ISSUER}) {
			return nil, errors.New("Invalid role")
		}
		return t.findConfirmedTransactionBySymbol(stub, args)
	} else if function == "getMoney" {
		if !t.stringInSlice(role, []string{ROLE_TRADER, ROLE_ISSUER}) {
			return nil, errors.New("Invalid role")
		}
		return t.getMoney(stub, args)
	} else if function == "getMaxNumberHolder" {
		if !t.stringInSlice(role, []string{ROLE_TRADER, ROLE_ISSUER}) {
			return nil, errors.New("Invalid role")
		}
		return t.getMaxNumberHolder(stub, args)
	} else if function == "getHolders" {
		if !t.stringInSlice(role, []string{ROLE_ISSUER}) {
			return nil, errors.New("Invalid role")
		}
		return t.getHolders(stub, args)
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
