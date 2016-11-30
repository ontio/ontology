package ledger


// Store provides storage for State data
type BlockchainStore interface {
	//TODO: define the state store func
	SaveBlock(*Block) error
}




type Blockchain struct {
	Store BlockchainStore
	TxPool TranscationPool
}