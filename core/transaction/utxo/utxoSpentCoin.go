package utxo

type ISpentCoin struct {
	Output      *TxOutput
	StartHeight uint32
	EndHeight   uint32
	Value       uint32
}
