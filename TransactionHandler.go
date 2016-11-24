package main

import (
  "encoding/json"
  "errors"
  "strconv"
  "time"

  "github.com/hyperledger/fabric/core/chaincode/shim"
)

const (
  tableTransaction    = "Transaction"
  columnTransactionID = "TransactionID"
  columnSymbol        = "Symbol"
  columnBuyerID       = "BuyerID"
  columnSellerID      = "SellerID"
  columnPrice         = "Price"
  columnVolume        = "Volume"
  columnStatus        = "Status"
  columnLastUpdated   = "LastUpdated"

  tableAccountIDTransaction = "AccountIDTx"
  columnAccountID           = "AccountID"

  stateCurrTransactionID = "CurrTransactionID"

  STATUS_WAITING   = "Waiting"
  STATUS_CONFIRMED = "Complete"
  STATUS_CANCEL_BUYER = "Cancelled By Buyer"
  STATUS_CANCEL_SELLER = "Cancelled By Seller"
)

type transactionHandler struct {
}

type TransactionMsg struct {
  TransactionID uint64
  Symbol string
  BuyerID string
  SellerID string
  Price string
  Volume uint64
  Status string
  LastUpdated string
}

func NewTransactionHandler() *transactionHandler {
  return &transactionHandler{}
}

func (t *transactionHandler) getCurrentTime() string {
  return time.Now().Format(time.RFC3339Nano)
}

func (t *transactionHandler) createTable(stub shim.ChaincodeStubInterface) error {
  stub.CreateTable(tableTransaction, []*shim.ColumnDefinition{
    &shim.ColumnDefinition{Name: columnTransactionID, Type: shim.ColumnDefinition_UINT64, Key: true},
    &shim.ColumnDefinition{Name: columnSymbol, Type: shim.ColumnDefinition_STRING, Key: false},
    &shim.ColumnDefinition{Name: columnBuyerID, Type: shim.ColumnDefinition_STRING, Key: false},
    &shim.ColumnDefinition{Name: columnSellerID, Type: shim.ColumnDefinition_STRING, Key: false},
    // &shim.ColumnDefinition{Name: columnPrice, Type: shim.ColumnDefinition_BYTES, Key: false},
    &shim.ColumnDefinition{Name: columnPrice, Type: shim.ColumnDefinition_STRING, Key: false},
    &shim.ColumnDefinition{Name: columnVolume, Type: shim.ColumnDefinition_UINT64, Key: false},
    &shim.ColumnDefinition{Name: columnStatus, Type: shim.ColumnDefinition_STRING, Key: false},
    &shim.ColumnDefinition{Name: columnLastUpdated, Type: shim.ColumnDefinition_STRING, Key: false},
  })

  stub.CreateTable(tableAccountIDTransaction, []*shim.ColumnDefinition{
    &shim.ColumnDefinition{Name: columnAccountID, Type: shim.ColumnDefinition_STRING, Key: true},
    &shim.ColumnDefinition{Name: columnSymbol, Type: shim.ColumnDefinition_STRING, Key: true},
    &shim.ColumnDefinition{Name: columnTransactionID, Type: shim.ColumnDefinition_UINT64, Key: true},
  })

  return nil
}

func (t *transactionHandler) insert(stub shim.ChaincodeStubInterface,
  symbol string,
  buyerID string,
  sellerID string,
  price string,
  // price []byte,
  volume uint64,
  status string) error {

  var tmpTxID int
  var txID uint64

  tmpbytes, err := stub.GetState(stateCurrTransactionID)
  if err != nil || tmpbytes == nil {
    txID = 1
  } else {
    tmpTxID, _ = strconv.Atoi(string(tmpbytes))
    txID = uint64(tmpTxID)
    txID++
  }
  err = stub.PutState(stateCurrTransactionID, []byte(strconv.FormatUint(txID, 10)))

  myLogger.Debugf("insert transactionID= %v", txID)

  ok, err := stub.InsertRow(tableTransaction, shim.Row{
    Columns: []*shim.Column{
      &shim.Column{Value: &shim.Column_Uint64{Uint64: txID}},
      &shim.Column{Value: &shim.Column_String_{String_: symbol}},
      &shim.Column{Value: &shim.Column_String_{String_: buyerID}},
      &shim.Column{Value: &shim.Column_String_{String_: sellerID}},
      // &shim.Column{Value: &shim.Column_Bytes{Bytes: price}},
      &shim.Column{Value: &shim.Column_String_{String_: price}},
      &shim.Column{Value: &shim.Column_Uint64{Uint64: volume}},
      &shim.Column{Value: &shim.Column_String_{String_: status}},
      &shim.Column{Value: &shim.Column_String_{String_: t.getCurrentTime()}}},
  })

  if !ok && err == nil {
    myLogger.Errorf("system error %v", err)
    return errors.New("Cannot insert transaction.")
  }

  ok, err = stub.InsertRow(tableAccountIDTransaction, shim.Row{
    Columns: []*shim.Column{
      &shim.Column{Value: &shim.Column_String_{String_: buyerID}},
      &shim.Column{Value: &shim.Column_String_{String_: symbol}},
      &shim.Column{Value: &shim.Column_Uint64{Uint64: txID}}},
  })

  if !ok && err == nil {
    myLogger.Errorf("system error %v", err)
    return errors.New("Cannot insert transaction.")
  }

  ok, err = stub.InsertRow(tableAccountIDTransaction, shim.Row{
    Columns: []*shim.Column{
      &shim.Column{Value: &shim.Column_String_{String_: sellerID}},
      &shim.Column{Value: &shim.Column_String_{String_: symbol}},
      &shim.Column{Value: &shim.Column_Uint64{Uint64: txID}}},
  })

  if !ok && err == nil {
    myLogger.Errorf("system error %v", err)
    return errors.New("Cannot insert transaction.")
  }

  return nil
}

func (t *transactionHandler) updateStatus(stub shim.ChaincodeStubInterface,
  txID uint64,
  status string) error {

  var columnsTx []shim.Column
  colTxId := shim.Column{Value: &shim.Column_Uint64{Uint64: txID}}
  columnsTx = append(columnsTx, colTxId)
  row, err := stub.GetRow(tableTransaction, columnsTx)

  if err != nil {
    myLogger.Errorf("system error %v", err)
    return errors.New("Cannot update transaction.")
  }

  if len(row.Columns) == 0 {
		return errors.New("Cannot update transaction.")
	}

  ok, err := stub.ReplaceRow(tableTransaction, shim.Row{
		Columns: []*shim.Column{
      &shim.Column{Value: &shim.Column_Uint64{Uint64: txID}},
      &shim.Column{Value: &shim.Column_String_{String_: row.Columns[1].GetString_()}},//symbol
      &shim.Column{Value: &shim.Column_String_{String_: row.Columns[2].GetString_()}},//buyerID
      &shim.Column{Value: &shim.Column_String_{String_: row.Columns[3].GetString_()}},//sellerID
      // &shim.Column{Value: &shim.Column_Bytes{Bytes: price}},
      &shim.Column{Value: &shim.Column_String_{String_: row.Columns[4].GetString_()}},//price
      &shim.Column{Value: &shim.Column_Uint64{Uint64: row.Columns[5].GetUint64()}},//volume
      &shim.Column{Value: &shim.Column_String_{String_: status}},//status
      &shim.Column{Value: &shim.Column_String_{String_: t.getCurrentTime()}}},
	})

	if !ok && err == nil {
		myLogger.Errorf("system error %v", err)
		return errors.New("Cannot update transaction.")
	}

  return nil
}

func (t *transactionHandler) getTransaction(stub shim.ChaincodeStubInterface,
  txID uint64) (*TransactionMsg, error) {

  var columnsTx []shim.Column
  colTxId := shim.Column{Value: &shim.Column_Uint64{Uint64: txID}}
  columnsTx = append(columnsTx, colTxId)
  row, err := stub.GetRow(tableTransaction, columnsTx)

  if err != nil {
    myLogger.Errorf("system error %v", err)
    return nil, errors.New("Cannot get transaction.")
  }

  if len(row.Columns) == 0 {
		return nil, nil
	}

  txMsg := TransactionMsg{
    row.Columns[0].GetUint64(),//txId
    row.Columns[1].GetString_(),//symbol
    row.Columns[2].GetString_(),//buyerID
    row.Columns[3].GetString_(),//sellerID
    row.Columns[4].GetString_(),//price
    row.Columns[5].GetUint64(),//volume
    row.Columns[6].GetString_(),//status
    row.Columns[7].GetString_(),//lastUpdated
  }

  return &txMsg, nil
}

func (t *transactionHandler) findTransactionByAccountID(stub shim.ChaincodeStubInterface,
  accountid string) ([]TransactionMsg, error) {

  var columns []shim.Column
  colAccountID := shim.Column{Value: &shim.Column_String_{String_: accountid}}
  columns = append(columns, colAccountID)

  rowChannel, err := stub.GetRows(tableAccountIDTransaction, columns)
  if err != nil {
    myLogger.Errorf("system error %v", err)
    return nil, errors.New("Cannot query transaction.")
  }

  var txMsgs []TransactionMsg

  for {
    select {
    case row, ok := <-rowChannel:
      if !ok {
        rowChannel = nil
      } else {
        var columnsTx []shim.Column
        colTxId := shim.Column{Value: &shim.Column_Uint64{Uint64: row.Columns[2].GetUint64()}}
        columnsTx = append(columnsTx, colTxId)
        rowTx, err := stub.GetRow(tableTransaction, columnsTx)

        if err != nil {
          myLogger.Errorf("system error %v", err)
          return nil, errors.New("Cannot query transaction.")
        }

        txMsg := TransactionMsg{
          rowTx.Columns[0].GetUint64(),//txId
          rowTx.Columns[1].GetString_(),//symbol
          rowTx.Columns[2].GetString_(),//buyerID
          rowTx.Columns[3].GetString_(),//sellerID
          rowTx.Columns[4].GetString_(),//price
          rowTx.Columns[5].GetUint64(),//volume
          rowTx.Columns[6].GetString_(),//status
          rowTx.Columns[7].GetString_(),//lastUpdated
        }
        txMsgs = append(txMsgs, txMsg)

        myLogger.Debugf("[%v]", txMsg)
      }
    }
    if rowChannel == nil {
      break
    }
  }

  return txMsgs, nil
}

func (t *transactionHandler) findTransactionByAccountIDSymbol(stub shim.ChaincodeStubInterface,
  accountid string,
  symbol string) ([]TransactionMsg, error) {

  var columns []shim.Column
  colAccountID := shim.Column{Value: &shim.Column_String_{String_: accountid}}
  colSymbol := shim.Column{Value: &shim.Column_String_{String_: symbol}}
  columns = append(columns, colAccountID)
  columns = append(columns, colSymbol)

  rowChannel, err := stub.GetRows(tableAccountIDTransaction, columns)
  if err != nil {
    myLogger.Errorf("system error %v", err)
    return nil, errors.New("Cannot query transaction.")
  }

  var txMsgs []TransactionMsg

  for {
    select {
    case row, ok := <-rowChannel:
      if !ok {
        rowChannel = nil
      } else {
        var columnsTx []shim.Column
        colTxId := shim.Column{Value: &shim.Column_Uint64{Uint64: row.Columns[2].GetUint64()}}
        columnsTx = append(columnsTx, colTxId)
        rowTx, err := stub.GetRow(tableTransaction, columnsTx)

        if err != nil {
          myLogger.Errorf("system error %v", err)
          return nil, errors.New("Cannot query transaction.")
        }

        txMsg := TransactionMsg{
          rowTx.Columns[0].GetUint64(),//txId
          rowTx.Columns[1].GetString_(),//symbol
          rowTx.Columns[2].GetString_(),//buyerID
          rowTx.Columns[3].GetString_(),//sellerID
          rowTx.Columns[4].GetString_(),//price
          rowTx.Columns[5].GetUint64(),//volume
          rowTx.Columns[6].GetString_(),//status
          rowTx.Columns[7].GetString_(),//lastUpdated
        }
        txMsgs = append(txMsgs, txMsg)

        myLogger.Debugf("[%v]", txMsg)
      }
    }
    if rowChannel == nil {
      break
    }
  }

  return txMsgs, nil
}

func (t *transactionHandler) query(stub shim.ChaincodeStubInterface,
  accountid string) ([]byte, error) {

  var columns []shim.Column
  colAccountID := shim.Column{Value: &shim.Column_String_{String_: accountid}}
  columns = append(columns, colAccountID)

  rowChannel, err := stub.GetRows(tableAccountIDTransaction, columns)
  if err != nil {
    myLogger.Errorf("system error %v", err)
    return nil, errors.New("Cannot query transaction.")
  }

  var txMsgs []TransactionMsg

  for {
    select {
    case row, ok := <-rowChannel:
      if !ok {
        rowChannel = nil
      } else {
        var columnsTx []shim.Column
        colTxId := shim.Column{Value: &shim.Column_Uint64{Uint64: row.Columns[2].GetUint64()}}
        columnsTx = append(columnsTx, colTxId)
        rowTx, err := stub.GetRow(tableTransaction, columnsTx)

        if err != nil {
          myLogger.Errorf("system error %v", err)
          return nil, errors.New("Cannot query transaction.")
        }

        txMsg := TransactionMsg{
          rowTx.Columns[0].GetUint64(),//txId
          rowTx.Columns[1].GetString_(),//symbol
          rowTx.Columns[2].GetString_(),//buyerID
          rowTx.Columns[3].GetString_(),//sellerID
          rowTx.Columns[4].GetString_(),//price
          rowTx.Columns[5].GetUint64(),//volume
          rowTx.Columns[6].GetString_(),//status
          rowTx.Columns[7].GetString_(),//lastUpdated
        }
        txMsgs = append(txMsgs, txMsg)

        myLogger.Debugf("[%v]", txMsg)
      }
    }
    if rowChannel == nil {
      break
    }
  }

  txMsgsJson, err := json.Marshal(txMsgs)
  myLogger.Debugf("Response : %s",  txMsgsJson)

  return txMsgsJson, nil
}
