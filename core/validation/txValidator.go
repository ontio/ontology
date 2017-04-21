package validation

import (
	"DNA/common"
	"DNA/core/ledger"
	tx "DNA/core/transaction"
	"DNA/core/transaction/payload"
	"errors"
	"math"
)

//Verfiy the transcation for following points
//- Well format
//- No duplicated inputs
//- inputs/outputs balance
//- Transcation contracts pass
func VerifyTransaction(Tx *tx.Transaction, ledger *ledger.Ledger, TxPool []*tx.Transaction) error {

	err := CheckDuplicateInput(Tx)
	if err != nil {
		return err
	}

	err = IsDoubleSpend(Tx, ledger)
	if err != nil {
		return err
	}

	//this func is changed to can process by Goroutine ,so move out this func individually in below.
	//if CheckMemPool(Tx, TxPool) {
	//	return fmt.Errorf("There is duplicated Tx Input with Tx Pool.")
	//}

	err = CheckAssetPrecision(Tx)
	if err != nil {
		return err
	}

	err = CheckTransactionBalance(Tx)
	if err != nil {
		return err
	}

	err = CheckAttributeProgram(Tx)
	if err != nil {
		return err
	}

	err = CheckTransactionContracts(Tx)
	if err != nil {
		return err
	}

	if Tx.TxType == tx.IssueAsset {
		results, err := Tx.GetTransactionResults()
		if err != nil {
			return errors.New("[VerifyTransaction], GetTransactionResults failed.")
		}

		for _, v := range results {
			//Get the Asset amount when RegisterAsseted.
			trx, err := tx.TxStore.GetTransaction(v.AssetId)
			if err != nil {
				return errors.New("[VerifyTransaction], AssetId does exist.")
			}
			if trx.TxType != tx.RegisterAsset {
				return errors.New("[VerifyTransaction], TxType is illegal.")
			}
			AssetReg := trx.Payload.(*payload.RegisterAsset)

			//Get the amount has been issued of this assetID
			var quantity_issued common.Fixed64
			if AssetReg.Amount < common.Fixed64(0) {
				continue
			} else {
				quantity_issued, err = tx.TxStore.GetQuantityIssued(v.AssetId)
				if err != nil {
					return errors.New("[VerifyTransaction], GetQuantityIssued failed.")
				}
			}

			//calc the amounts in txPool
			var txPoolAmounts common.Fixed64
			for _, t := range TxPool {
				for _, outputs := range t.Outputs {
					if outputs.AssetID == v.AssetId {
						txPoolAmounts = txPoolAmounts + outputs.Value * -1
					}
				}
			}

			//calc weather out off the amount when Registed.
			//AssetReg.Amount : amount when RegisterAsset of this assedID
			//quantity_issued : amount has been issued of this assedID
			//txPoolAmounts   : amount in transactionPool of this assedID

			// TODO: check this after the function TxStore.GetQuantityIssued works
			if AssetReg.Amount-quantity_issued < txPoolAmounts {
				return errors.New("[VerifyTransaction], Amount check error.")
			}
		}
	}

	return nil
}


//Use for verify request, only response validate/invalidate.
func VerifyTransactionPoolWhenResponse (txPool  []*tx.Transaction)bool{
	if len(txPool) == 0 {
		return true
	}

	utxoMap := make(map[string]bool, 0)
	for _, t := range txPool{
		for _, u := range t.UTXOInputs{
			utxo := u.ToString()
			if _, ok := utxoMap[utxo]; ok {
				return false
			}else{
				utxoMap[utxo] = true
			}
		}
	}

	return true
}

//Use for request by miner.
//remove the invalidate transaction from process context and tell the caller which should be removed.
func VerifyTransactionPoolWhenRequest(txPool  map[common.Uint256]*tx.Transaction)(txs []common.Uint256,NewPool map[common.Uint256]*tx.Transaction){
	if len(txPool) == 0 {
		return nil,txPool
	}

	errorTxs :=  []common.Uint256{}
	utxoMap := make(map[string]common.Uint256, 0)
	for k, t := range txPool{
		for _, u := range t.UTXOInputs{
			utxo := u.ToString()
			if v, ok := utxoMap[utxo]; ok {
				delete(txPool,v)
				delete(txPool,k)
				errorTxs = append(errorTxs,k)
				errorTxs = append(errorTxs,v)
				continue
			}else{
				utxoMap[utxo] = k
			}
		}
	}

	return errorTxs,txPool
}

func CheckMemPool(tx *tx.Transaction, TxPool []*tx.Transaction) error {
	if len(tx.UTXOInputs) == 0 {
		return nil
	}
	for _, poolTx := range TxPool {
		for _, poolInput := range poolTx.UTXOInputs {
			for _, txInput := range tx.UTXOInputs {
				if poolInput.Equals(txInput) {
					return errors.New("There is duplicated Tx Input with Tx Pool.")
				}
			}
		}
	}
	return nil
}

func CheckDuplicateInput(tx *tx.Transaction) error {
	if len(tx.UTXOInputs) == 0 {
		return nil
	}
	for i, utxoin := range tx.UTXOInputs {
		for j := 0; j < i; j++ {
			if utxoin.ReferTxID == tx.UTXOInputs[j].ReferTxID && utxoin.ReferTxOutputIndex == tx.UTXOInputs[j].ReferTxOutputIndex {
				return errors.New("invalid transaction")
			}
		}
	}
	return nil
}

func IsDoubleSpend(tx *tx.Transaction, ledger *ledger.Ledger) error {
	return ledger.IsDoubleSpend(tx)
}

func CheckAssetPrecision(Tx *tx.Transaction) error {
	for k, outputs := range Tx.AssetOutputs {
		asset, err := ledger.DefaultLedger.GetAsset(k)
		if err != nil {
			return errors.New("The asset not exist in local blockchain.")
		}
		precision := asset.Precision
		for _, output := range outputs {
			if output.Value.GetData()%int64(math.Pow(10, 8-float64(precision))) != 0 {
				return errors.New("The precision of asset is incorrect.")
			}
		}
	}
	return nil
}

func CheckTransactionBalance(Tx *tx.Transaction) error {
	if len(Tx.AssetInputAmount) != len(Tx.AssetOutputAmount) {
		return errors.New("The number of asset is not same between inputs and outputs.")
	}

	for k, v := range Tx.AssetInputAmount {
		if v != Tx.AssetOutputAmount[k] {
			return errors.New("The amount of asset is not same between inputs and outputs.")
		}
	}
	return nil
}

func CheckAttributeProgram(Tx *tx.Transaction) error {
	//TODO: implement CheckAttributeProgram
	return nil
}

func CheckTransactionContracts(Tx *tx.Transaction) error {
	flag, err := VerifySignableData(Tx)
	if flag && err == nil {
		return nil
	} else {
		return err
	}
}
