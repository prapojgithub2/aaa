package main

import (
	"errors"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

const (
	tableAccountMoney = "AccountMoney"
	columnAmount      = "Amount"
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
	stub.CreateTable(tableAccountMoney, []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: columnAccountID, Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: columnAmount, Type: shim.ColumnDefinition_UINT64, Key: false},
	})
	return t.initAccountMoney(stub)
}

func (t *accountMoneyHandler) initAccountMoney(stub shim.ChaincodeStubInterface) error {
	t.assign(stub, "AA01", 100000)
	t.assign(stub, "AA02", 100000)
	t.assign(stub, "AA03", 100000)
	t.assign(stub, "AA04", 100000)
	t.assign(stub, "AA05", 100000)
	t.assign(stub, "AA06", 100000)
	t.assign(stub, "AA07", 100000)
	t.assign(stub, "AA08", 100000)
	t.assign(stub, "AA09", 100000)
	t.assign(stub, "AA10", 100000)
	return nil
}

func (t *accountMoneyHandler) assign(stub shim.ChaincodeStubInterface,
	accountID string,
	amount uint64) error {

	myLogger.Debugf("insert accountID= %v", accountID)

	//insert a new row for this account ID that includes contact information and balance
	ok, err := stub.InsertRow(tableAccountMoney, shim.Row{
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

	ok, err := stub.ReplaceRow(tableAccountMoney, shim.Row{
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
		"tableAccountMoney",
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

	return stub.GetRow(tableAccountMoney, columns)
}
