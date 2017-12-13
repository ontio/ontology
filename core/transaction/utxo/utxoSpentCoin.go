package utxo

type SpentCoin struct {
	Output      *TxOutput
	StartHeight uint32
	EndHeight   uint32
	Value       uint32
}