package ledger

import (
	. "DNA/common"
	"DNA/common/log"
	"DNA/common/serialization"
	"DNA/core/contract/program"
	sig "DNA/core/signature"
	tx "DNA/core/transaction"
	"DNA/crypto"
	. "DNA/errors"
	"DNA/vm"
	"io"
	"time"
)

type Block struct {
	Blockdata    *Blockdata
	Transactions []*tx.Transaction

	hash *Uint256
}

func (b *Block) Serialize(w io.Writer) error {
	b.Blockdata.Serialize(w)
	err := serialization.WriteUint32(w, uint32(len(b.Transactions)))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Block item Transactions length serialization failed.")
	}

	for _, transaction := range b.Transactions {
		transaction.Serialize(w)
	}
	return nil
}

func (b *Block) Deserialize(r io.Reader) error {
	if b.Blockdata == nil {
		b.Blockdata = new(Blockdata)
	}
	b.Blockdata.Deserialize(r)

	//Transactions
	var i uint32
	Len, err := serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	var txhash Uint256
	var tharray []Uint256
	for i = 0; i < Len; i++ {
		transaction := new(tx.Transaction)
		transaction.Deserialize(r)
		txhash = transaction.Hash()
		b.Transactions = append(b.Transactions, transaction)
		tharray = append(tharray, txhash)
	}

	b.Blockdata.TransactionsRoot, err = crypto.ComputeRoot(tharray)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Block Deserialize merkleTree compute failed")
	}

	return nil
}

func (b *Block) Trim(w io.Writer) error {
	b.Blockdata.Serialize(w)
	err := serialization.WriteUint32(w, uint32(len(b.Transactions)))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Block item Transactions length serialization failed.")
	}
	for _, transaction := range b.Transactions {
		temp := *transaction
		hash := temp.Hash()
		hash.Serialize(w)
	}
	return nil
}

func (b *Block) FromTrimmedData(r io.Reader) error {
	if b.Blockdata == nil {
		b.Blockdata = new(Blockdata)
	}
	b.Blockdata.Deserialize(r)

	//Transactions
	var i uint32
	Len, err := serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	var txhash Uint256
	var tharray []Uint256
	for i = 0; i < Len; i++ {
		txhash.Deserialize(r)
		transaction := new(tx.Transaction)
		transaction.SetHash(txhash)
		b.Transactions = append(b.Transactions, transaction)
		tharray = append(tharray, txhash)
	}

	b.Blockdata.TransactionsRoot, err = crypto.ComputeRoot(tharray)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Block Deserialize merkleTree compute failed")
	}

	return nil
}

func (b *Block) GetMessage() []byte {
	return sig.GetHashForSigning(b)
}

func (b *Block) GetProgramHashes() ([]Uint160, error) {

	return b.Blockdata.GetProgramHashes()
}

func (b *Block) SetPrograms(prog []*program.Program) {
	b.Blockdata.SetPrograms(prog)
	return
}

func (b *Block) GetPrograms() []*program.Program {
	return b.Blockdata.GetPrograms()
}

func (b *Block) Hash() Uint256 {
	if b.hash == nil {
		b.hash = new(Uint256)
		*b.hash = b.Blockdata.Hash()
	}
	return *b.hash
}

func (b *Block) Verify() error {
	log.Info("This function is expired.please use Validation/blockValidator to Verify Block.")
	return nil
}

func (b *Block) Type() InventoryType {
	return BLOCK
}

func GenesisBlockInit() (*Block, error) {
	genesisBlock := new(Block)
	//blockdata
	genesisBlockdata := new(Blockdata)
	genesisBlockdata.Version = uint32(0x00)
	genesisBlockdata.PrevBlockHash = Uint256{}
	genesisBlockdata.TransactionsRoot = Uint256{}
	tm := time.Date(2017, time.February, 23, 0, 0, 0, 0, time.UTC)
	genesisBlockdata.Timestamp = uint32(tm.Unix())
	genesisBlockdata.Height = uint32(0)
	genesisBlockdata.ConsensusData = uint64(2083236893)
	nextBookKeeper, err := GetBookKeeperAddress(StandbyBookKeepers)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Block],GenesisBlockInit err with GetBookKeeperAddress")
	}
	genesisBlockdata.NextBookKeeper = nextBookKeeper

	pg := new(program.Program)
	pg.Code = []byte{'0'}
	pg.Parameter = []byte{byte(vm.PUSHT)}
	genesisBlockdata.Program = pg

	//transaction
	trans := new(tx.Transaction)
	{
		trans.TxType = tx.BookKeeping
		trans.PayloadVersion = byte(0)
		trans.Payload = nil
		trans.Nonce = uint64(0)
		trans.Attributes = nil
		trans.UTXOInputs = nil
		trans.BalanceInputs = nil
		trans.Outputs = nil
		{
			programHashes := []*program.Program{}
			pg := new(program.Program)
			pg.Code = []byte{'0'}
			pg.Parameter = []byte{byte(vm.PUSHT)}
			programHashes = append(programHashes, pg)
			trans.Programs = programHashes
		}
	}
	genesisBlock.Blockdata = genesisBlockdata

	genesisBlock.Transactions = append(genesisBlock.Transactions, trans)

	//hashx := genesisBlock.Hash()

	return genesisBlock, nil
}

func (b *Block) RebuildMerkleRoot() error {
	txs := b.Transactions
	transactionHashes := []Uint256{}
	for _, tx := range txs {
		transactionHashes = append(transactionHashes, tx.Hash())
	}
	hash, err := crypto.ComputeRoot(transactionHashes)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[Block] , RebuildMerkleRoot ComputeRoot failed.")
	}
	b.Blockdata.TransactionsRoot = hash
	return nil

}

func (bd *Block) SerializeUnsigned(w io.Writer) error {
	return bd.Blockdata.SerializeUnsigned(w)
}
