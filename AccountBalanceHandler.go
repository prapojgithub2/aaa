package main

import (
  "encoding/json"
  "errors"
  "github.com/hyperledger/fabric/core/chaincode/shim"
)

const (
  tableAccountBalance = "AccountBalance"
  columnBalance       = "Balance"
)

type accountBalanceHandler struct {
}

type BalanceMsg struct {
  AccountID string
  Symbol string
  Balance uint64
}

func NewAccountBalanceHandler() *accountBalanceHandler {
  return &accountBalanceHandler{}
}

func (t *accountBalanceHandler) createTable(stub shim.ChaincodeStubInterface) error {

  err := stub.CreateTable(tableAccountBalance, []*shim.ColumnDefinition{
    &shim.ColumnDefinition{Name: columnAccountID, Type: shim.ColumnDefinition_STRING, Key: true},
    &shim.ColumnDefinition{Name: columnSymbol, Type: shim.ColumnDefinition_STRING, Key: true},
    &shim.ColumnDefinition{Name: columnBalance, Type: shim.ColumnDefinition_UINT64, Key: false},
  })

  if err != nil {
    myLogger.Errorf("system error create table account balance %v", err)
    return errors.New("Cannot create table account balance.")
  }

  return t.InitAccountBalance(stub);

}

func (t *accountBalanceHandler) listHolderBySymbol(stub shim.ChaincodeStubInterface,symbol string) ([]byte,error) {
  var balMsgs []BalanceMsg 
  var columnsTx []shim.Column

  rowChannel, err := stub.GetRows(tableAccountBalance, columnsTx)
  if err != nil {
    myLogger.Errorf("system error %v", err)
    return nil, errors.New("Cannot query account balance.")
  }
  
  for {
    select {
    case row, ok := <-rowChannel:
      if !ok {
        rowChannel = nil
      } else {
            tAccount := row.Columns[0].GetString_();
            tSymbol := row.Columns[1].GetString_();
            tBalance := row.Columns[2].GetUint64();

            if (tSymbol == symbol && tBalance > 0){ 
              balMsg := BalanceMsg{
                tAccount,
                tSymbol,
                tBalance,
              }
            balMsgs = append(balMsgs, balMsg)
          }
      }
      if rowChannel == nil {
        break
      }
    }
  }

  balMsgsJson, err := json.Marshal(balMsgs)
  myLogger.Debugf("Response : %s",  balMsgsJson)
  return balMsgsJson, nil
}


func (t *accountBalanceHandler) InitAccountBalance(stub shim.ChaincodeStubInterface) error {
	  t.updateAccountBalance(stub,"AA01","AAAA", 1000)
  	t.updateAccountBalance(stub,"AA01","BBBB", 1000)
  	t.updateAccountBalance(stub,"AA01","CCCC", 1000)
  	t.updateAccountBalance(stub,"AA01","DDDD", 1000)
	  t.updateAccountBalance(stub,"AA02","AAAA", 1000)
  	t.updateAccountBalance(stub,"AA02","BBBB", 1000)
  	t.updateAccountBalance(stub,"AA02","CCCC", 1000)
  	t.updateAccountBalance(stub,"AA02","DDDD", 1000)
  	t.updateAccountBalance(stub,"AA03","AAAA", 1000)
  	t.updateAccountBalance(stub,"AA03","BBBB", 1000)
  	t.updateAccountBalance(stub,"AA03","CCCC", 1000)
  	t.updateAccountBalance(stub,"AA03","DDDD", 1000)
  	t.updateAccountBalance(stub,"AA04","AAAA", 1000)
  	t.updateAccountBalance(stub,"AA04","BBBB", 1000)
  	t.updateAccountBalance(stub,"AA04","CCCC", 1000)
  	t.updateAccountBalance(stub,"AA04","DDDD", 1000)  	
   	t.updateAccountBalance(stub,"AA05","AAAA", 1000)
  	t.updateAccountBalance(stub,"AA05","BBBB", 1000)
  	t.updateAccountBalance(stub,"AA05","CCCC", 1000)
  	t.updateAccountBalance(stub,"AA05","DDDD", 1000)  
  return nil
}

func (t *accountBalanceHandler) validateOverTermSheetRules(stub shim.ChaincodeStubInterface,
  sellerID string, 
  buyerID string,
  symbol string, 
  volume uint64,
  noOfHolderAllowed uint64) (bool,error){

  var columnsTx []shim.Column
  rowChannel, err := stub.GetRows(tableAccountBalance, columnsTx)
  if err != nil {
    myLogger.Errorf("system error %v", err)
    return false, errors.New("Cannot query account balance.")
  }
  var finalNoOfHolders uint64 = 1; // for buyers
  validSeller := false;
  overRide := false;
  
  for {
    select {
    case row, ok := <-rowChannel:
      if !ok {
        rowChannel = nil
      } else {
            tAccount := row.Columns[0].GetString_();
            tSymbol := row.Columns[1].GetString_();
            tBalance := row.Columns[2].GetUint64();
            /*symbol*/ 
            if (tSymbol == symbol ){ 
              if ( tBalance > 0 ) { finalNoOfHolders = finalNoOfHolders + 1 }
              /*account check for seller*/ 
              if (tAccount == sellerID && tBalance >= volume){
                validSeller = true;
                myLogger.Infof("++++++++++++++++++++++++ Can sell [%v,%v]",sellerID,tBalance);
                if (tBalance == volume) {                
                  finalNoOfHolders = finalNoOfHolders - 1;                
                }
              }

              /*account check for buyer*/ 
              if (tAccount == buyerID && tBalance > 0){
                myLogger.Infof("++++++++++++++++++++++++ Sell to existing holders [%v,%v]",buyerID,tBalance);
                overRide = true;
                finalNoOfHolders = finalNoOfHolders - 1;                
              }
            }
      }  
    }
    if rowChannel == nil {
      break
    }
  }
  myLogger.Infof("++++++++++++++ validateOverTermSheetRules overRide=%v,NoOfHolders=%v,ValidSeller=%v",overRide,finalNoOfHolders,validSeller);
  if ( (overRide || finalNoOfHolders <= noOfHolderAllowed) && validSeller){
      return true , nil
  } 
  return false , nil;
}



func (t *accountBalanceHandler) updateAccountBalance(stub shim.ChaincodeStubInterface,
  accountID string, 
  symbol string,
  balance uint64) error {

  var columnsTx []shim.Column
  colAccountID := shim.Column{Value: &shim.Column_String_{String_: accountID}}
  columnsTx = append(columnsTx, colAccountID)
  colSymbol := shim.Column{Value: &shim.Column_String_{String_: symbol}}
  columnsTx = append(columnsTx, colSymbol)
  row, err := stub.GetRow(tableAccountBalance, columnsTx)

  if err != nil {
    myLogger.Errorf("system error %v", err)
    return errors.New("Cannot update transaction.")
  }

  if len(row.Columns) == 0 {
  	_, err := stub.InsertRow(tableAccountBalance, shim.Row{
      Columns: []*shim.Column{
        &shim.Column{Value: &shim.Column_String_{String_: accountID}},
        &shim.Column{Value: &shim.Column_String_{String_: symbol}},
        &shim.Column{Value: &shim.Column_Uint64{Uint64: balance}}},  
    })

  	if err != nil {
    	myLogger.Errorf("system error %v", err)
    	return errors.New("Cannot insert account balance.")
  	}
  	return nil
  }

   _, err = stub.ReplaceRow(tableAccountBalance, shim.Row{
    Columns: []*shim.Column{
      &shim.Column{Value: &shim.Column_String_{String_: row.Columns[0].GetString_()}},//accountID
      &shim.Column{Value: &shim.Column_String_{String_: row.Columns[1].GetString_()}},//symbol
      &shim.Column{Value: &shim.Column_Uint64{Uint64: balance}}},//balance
  })

  if err != nil {
    myLogger.Errorf("system error %v", err)
    return errors.New("Cannot update account balance.")
  }

  return nil
}


func (t *accountBalanceHandler) query(stub shim.ChaincodeStubInterface, accountID string) ([]byte, error) {
  var columnsTx []shim.Column
  colAccountID := shim.Column{Value: &shim.Column_String_{String_: accountID}}
  columnsTx = append(columnsTx, colAccountID)
  rowChannel, err := stub.GetRows(tableAccountBalance, columnsTx)
  var balMsgs []BalanceMsg
  
  if err != nil {
    myLogger.Errorf("system error %v", err)
    return nil, errors.New("Cannot query account balance.")
  }

  
  for {
    select {
    case row, ok := <-rowChannel:
      if !ok {
        rowChannel = nil
      } else {

        balMsg := BalanceMsg{
          row.Columns[0].GetString_(),//accountID
          row.Columns[1].GetString_(),//symbol
          row.Columns[2].GetUint64(),//balance
        }
        balMsgs = append(balMsgs, balMsg)

        myLogger.Debugf("[%v]", balMsg)
      }
    }
    if rowChannel == nil {
      break
    }
  }
  balMsgsJson, err := json.Marshal(balMsgs)
  myLogger.Debugf("Response : %s",  balMsgsJson)

  return balMsgsJson, nil
}

func (t *accountBalanceHandler) transferAccountBalance(stub shim.ChaincodeStubInterface, sellerID string, buyerID string,symbol string, volume uint64) error {
  myLogger.Debugf("transfer balance %v , %v , %v, %v",sellerID,buyerID,symbol,volume)
  
  // seller
  var sellerColumnsTx []shim.Column
  colAccountID := shim.Column{Value: &shim.Column_String_{String_: sellerID}}
  sellerColumnsTx = append(sellerColumnsTx, colAccountID)
  colSymbol := shim.Column{Value: &shim.Column_String_{String_: symbol}}
  sellerColumnsTx = append(sellerColumnsTx, colSymbol)


  seller, err := stub.GetRow(tableAccountBalance, sellerColumnsTx)
  
  if err != nil  || len(seller.Columns) == 0 {
    myLogger.Errorf("system error %v", err)
    return errors.New("Cannot transfer account balance on sell side.")
  }

  myLogger.Debugf("**************transfer balance %v ",seller)
  if seller.Columns[2].GetUint64() < volume {
    myLogger.Error("query error not enough balance")
    return errors.New("Cannot transfer account balance on sell side.")
  }

  // buyer
  var buyerColumnsTx []shim.Column
  colAccountID = shim.Column{Value: &shim.Column_String_{String_: buyerID}}
  buyerColumnsTx = append(buyerColumnsTx, colAccountID)
  colSymbol = shim.Column{Value: &shim.Column_String_{String_: symbol}}
  buyerColumnsTx = append(buyerColumnsTx, colSymbol)

  buyer, err := stub.GetRow(tableAccountBalance, buyerColumnsTx)
  if err != nil  {
    myLogger.Errorf("system error %v", err)
    return errors.New("Cannot transfer account balance on buy side.")
  }
  var buyerBal uint64 = 0;
  if len(buyer.Columns) > 0 {
    buyerBal = buyer.Columns[2].GetUint64()
  }
  myLogger.Infof("+++++++++++++++++++   BuyerBal %v" , buyerBal)
  buyerBal = volume + buyerBal;
  t.updateAccountBalance(stub,sellerID,symbol,seller.Columns[2].GetUint64() - volume)
  t.updateAccountBalance(stub,buyerID,symbol,buyerBal)
  return nil
}
