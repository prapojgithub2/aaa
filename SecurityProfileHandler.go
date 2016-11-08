package main

import (
	"errors"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

const (
	tableSecurityProfile = "SecurityProfile"

	columnMaxNumberHolder = "MaxNumberHolder"
)

type securityProfileHandler struct {
}

//
type SecurityProfileMsg struct {
	Symbol          string
	MaxNumberHolder uint64
}

//
func NewSecurityProfileHandler() *securityProfileHandler {
	return &securityProfileHandler{}
}

func (t *securityProfileHandler) createTable(stub shim.ChaincodeStubInterface) error {

	// Create asset depository table
	stub.CreateTable(tableSecurityProfile, []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: columnSymbol, Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: columnMaxNumberHolder, Type: shim.ColumnDefinition_UINT64, Key: false},
	})
	return t.initSecurityProfile(stub)
}

func (t *securityProfileHandler) initSecurityProfile(stub shim.ChaincodeStubInterface) error {
	t.createSecurityProfile(stub, "AAAA", 10)
	t.createSecurityProfile(stub, "BBBB", 10)
	t.createSecurityProfile(stub, "CCCC", 10)
	t.createSecurityProfile(stub, "DDDD", 10)
	t.createSecurityProfile(stub, "EEEE", 10)
	return nil
}

func (t *securityProfileHandler) createSecurityProfile(stub shim.ChaincodeStubInterface,
	symbol string,
	maxNumberHolder uint64) error {

	myLogger.Debugf("insert symbol= %v", symbol)

	//insert a new row for this account ID that includes contact information and balance
	ok, err := stub.InsertRow(tableSecurityProfile, shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: symbol}},
			&shim.Column{Value: &shim.Column_Uint64{Uint64: maxNumberHolder}}},
	})

	// you can only assign balances to new account IDs
	if !ok && err == nil {
		myLogger.Errorf("system error %v", err)
		return errors.New("Asset was already assigned.")
	}

	return nil
}

func (t *securityProfileHandler) updateSecurityProfile(stub shim.ChaincodeStubInterface,
	symbol string,
	maxNumberHolder uint64) error {

	myLogger.Debugf("update symbol= %v", symbol)

	ok, err := stub.ReplaceRow(tableSecurityProfile, shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: symbol}},
			&shim.Column{Value: &shim.Column_Uint64{Uint64: maxNumberHolder}}},
	})

	if !ok && err == nil {
		myLogger.Errorf("system error %v", err)
		return errors.New("failed to replace row with symbol" + symbol)
	}
	return nil
}

func (t *securityProfileHandler) deleteAccountRecord(stub shim.ChaincodeStubInterface, symbol string) error {

	myLogger.Debugf("delete symbol= %v", symbol)

	//delete record matching account ID passed in
	err := stub.DeleteRow(
		tableSecurityProfile,
		[]shim.Column{shim.Column{Value: &shim.Column_String_{String_: symbol}}},
	)

	if err != nil {
		myLogger.Errorf("system error %v", err)
		return errors.New("error in deleting account record")
	}
	return nil
}

func (t *securityProfileHandler) getMaxNumberHolder(stub shim.ChaincodeStubInterface, symbol string) (uint64, error) {

	myLogger.Debugf("get Balance symbol= %v", symbol)

	row, err := t.queryTable(stub, symbol)
	if err != nil {
		return 0, err
	}
	if len(row.Columns) == 0 || row.Columns[1] == nil {
		return 0, errors.New("row or column value not found")
	}

	return row.Columns[1].GetUint64(), nil
}

func (t *securityProfileHandler) queryTable(stub shim.ChaincodeStubInterface, symbol string) (shim.Row, error) {

	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: symbol}}
	columns = append(columns, col1)

	return stub.GetRow(tableSecurityProfile, columns)
}
