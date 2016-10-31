package main

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

const (
	tableColumn     = "AccountMoney"
	columnAccountID = "AccountID"
	columnAmount    = "Amount"
)

type accountMoneyHandler struct {
}

//
type AccountMoneyMsg struct {
	AccountID string
	Amount    uint64
}

//
func NewAccountMoneyHandler() *accountMoneyHandler {
	return &accountMoneyHandler{}
}

func (t *accountMoneyHandler) createTable(stub shim.ChaincodeStubInterface) error {

	// Create asset depository table
	stub.CreateTable(tableColumn, []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: columnAccountID, Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: columnAmount, Type: shim.ColumnDefinition_UINT64, Key: false},
	})
	return t.initAccountMoney(stub)
}

func (t *accountMoneyHandler) initAccountMoney(stub shim.ChaincodeStubInterface) error {
	t.assign(stub, "A01", 100000)
	t.assign(stub, "A02", 100000)
	t.assign(stub, "A03", 100000)
	t.assign(stub, "A04", 100000)
	t.assign(stub, "A05", 100000)
	t.assign(stub, "A06", 100000)
	t.assign(stub, "A07", 100000)
	t.assign(stub, "A08", 100000)
	t.assign(stub, "A09", 100000)
	t.assign(stub, "A10", 100000)
	return nil
}

func (t *accountMoneyHandler) assign(stub shim.ChaincodeStubInterface,
	accountID string,
	amount uint64) error {

	myLogger.Debugf("insert accountID= %v", accountID)

	//insert a new row for this account ID that includes contact information and balance
	ok, err := stub.InsertRow(tableColumn, shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: accountID}},
			&shim.Column{Value: &shim.Column_Uint64{Uint64: amount}}},
	})

	// you can only assign balances to new account IDs
	if !ok && err == nil {
		myLogger.Errorf("system error %v", err)
		return errors.New("Asset was already assigned.")
	}

	return nil
}

func (t *accountMoneyHandler) updateAccountBalance(stub shim.ChaincodeStubInterface,
	accountID string,
	amount uint64) error {

	myLogger.Debugf("update accountID= %v", accountID)

	ok, err := stub.ReplaceRow(tableColumn, shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: accountID}},
			&shim.Column{Value: &shim.Column_Uint64{Uint64: amount}}},
	})

	if !ok && err == nil {
		myLogger.Errorf("system error %v", err)
		return errors.New("failed to replace row with account Id." + accountID)
	}
	return nil
}

func (t *accountMoneyHandler) deleteAccountRecord(stub shim.ChaincodeStubInterface, accountID string) error {

	myLogger.Debugf("delete accountID= %v", accountID)

	//delete record matching account ID passed in
	err := stub.DeleteRow(
		"AssetsOwnership",
		[]shim.Column{shim.Column{Value: &shim.Column_String_{String_: accountID}}},
	)

	if err != nil {
		myLogger.Errorf("system error %v", err)
		return errors.New("error in deleting account record")
	}
	return nil
}

func (t *accountMoneyHandler) transfer(stub shim.ChaincodeStubInterface, fromAccount string, toAccount string, amount uint64) error {

	myLogger.Debugf("transfer params= %v , %v , %v ", fromAccount, toAccount, amount)

	//collecting assets need to be transfered
	remaining := amount

	acctBalanceF, err := t.queryBalance(stub, fromAccount)
	if err != nil {
		myLogger.Errorf("system error %v", err)
		return errors.New("error in transfer money get acctBalanceF")
	}
	//check if toAccount
	acctBalanceT, err := t.queryBalance(stub, toAccount)
	if err != nil {
		myLogger.Errorf("system error %v", err)
		return errors.New("error in transfer money get acctBalanceT")
	}

	myLogger.Debugf("transfer acctBalanceF= %v", acctBalanceF)
	myLogger.Debugf("transfer acctBalanceT= %v", acctBalanceT)

	if remaining > 0 {

		if remaining > acctBalanceF {
			return errors.New("not enough money to transfer")
		}

		acctBalanceF -= remaining
		err = t.updateAccountBalance(stub, fromAccount, acctBalanceF)
		if err != nil {
			return errors.New("error in transfer money to fromAccount ")
		}

		acctBalanceT += remaining
		err = t.updateAccountBalance(stub, toAccount, acctBalanceT)
		if err != nil {
			return errors.New("error in transfer money to toAccount ")
		}

	}

	return nil
}

func (t *accountMoneyHandler) queryBalance(stub shim.ChaincodeStubInterface, accountID string) (uint64, error) {

	myLogger.Debugf("get Balance accountID= %v", accountID)

	row, err := t.queryTable(stub, accountID)
	if err != nil {
		return 0, err
	}
	if len(row.Columns) == 0 || row.Columns[1] == nil {
		return 0, errors.New("row or column value not found")
	}

	return row.Columns[1].GetUint64(), nil
}

func (t *accountMoneyHandler) queryTable(stub shim.ChaincodeStubInterface, accountID string) (shim.Row, error) {

	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: accountID}}
	columns = append(columns, col1)

	return stub.GetRow(tableColumn, columns)
}

func (t *accountMoneyHandler) getMoney(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debugf("+++++++++++++++++++++++++++++++++++getMoney+++++++++++++++++++++++++++++++++")
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}

	accountID := args[0]

	balance, err := t.queryBalance(stub, accountID)
	if err != nil {
		return nil, err
	}

	var txMsgs []AccountMoneyMsg
	txMsg := AccountMoneyMsg{
		accountID, //AccountID
		balance,   //Amount
	}
	txMsgs = append(txMsgs, txMsg)

	txMsgsJSON, err := json.Marshal(txMsgs)
	myLogger.Debugf("Response : %s", txMsgsJSON)

	//return []byte(balanceStr), nil
	return txMsgsJSON, nil
}

func (t *accountMoneyHandler) transferMoney(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debugf("+++++++++++++++++++++++++++++++++++transferMoney+++++++++++++++++++++++++++++++++")

	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}

	amount, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		myLogger.Errorf("system error %v", err)
		return nil, errors.New("Unable to parse amount" + args[2])
	}

	// call dHandler.transfer to transfer to transfer "amount" from "from account" IDs to "to account" IDs
	return nil, t.transfer(stub, args[0], args[1], amount)
}
