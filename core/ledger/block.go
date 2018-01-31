package ledger

import (
	"bytes"
	. "github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/core/contract/program"
	sig "github.com/Ontology/core/signature"
	tx "github.com/Ontology/core/transaction"
	"github.com/Ontology/core/transaction/payload"
	"github.com/Ontology/core/transaction/utxo"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/errors"
	vm "github.com/Ontology/vm/neovm"
	"io"
	"time"
)

const (
	BlockVersion      uint32 = 0
	GenesisNonce      uint64 = 2083236893
	DecrementInterval uint32 = 2000000
)

var (
	GenerationAmount = [17]uint32{80, 70, 60, 50, 40, 30, 20, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10}
)

var GenBlockTime = (config.DEFAULTGENBLOCKTIME * time.Second)

type Block struct {
	Header       *Header
	Transactions []*tx.Transaction

	hash *Uint256
}

func (b *Block) Serialize(w io.Writer) error {
	b.Header.Serialize(w)
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
	if b.Header == nil {
		b.Header = new(Header)
	}
	b.Header.Deserialize(r)

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

	b.Header.TransactionsRoot, err = crypto.ComputeRoot(tharray)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Block Deserialize merkleTree compute failed")
	}

	return nil
}

func (b *Block) Trim(w io.Writer) error {
	b.Header.Serialize(w)
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
	if b.Header == nil {
		b.Header = new(Header)
	}
	b.Header.Deserialize(r)

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

	b.Header.TransactionsRoot, err = crypto.ComputeRoot(tharray)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Block Deserialize merkleTree compute failed")
	}

	return nil
}

func (b *Block) GetMessage() []byte {
	return sig.GetHashData(b)
}

func (b *Block) GetProgramHashes() ([]Uint160, error) {

	return b.Header.GetProgramHashes()
}

func (b *Block) ToArray() []byte {
	bf := new(bytes.Buffer)
	b.Serialize(bf)
	return bf.Bytes()
}

func (b *Block) SetPrograms(prog []*program.Program) {
	b.Header.SetPrograms(prog)
	return
}

func (b *Block) GetPrograms() []*program.Program {
	return b.Header.GetPrograms()
}

func (b *Block) Hash() Uint256 {
	if b.hash == nil {
		b.hash = new(Uint256)
		*b.hash = b.Header.Hash()
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

func GenesisBlockInit(defaultBookKeeper []*crypto.PubKey) (*Block, error) {
	//getBookKeeper
	nextBookKeeper, err := GetBookKeeperAddress(defaultBookKeeper)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Block],GenesisBlockInit err with GetBookKeeperAddress")
	}
	//blockdata
	genesisHeader := &Header{
		Version:          BlockVersion,
		PrevBlockHash:    Uint256{},
		TransactionsRoot: Uint256{},
		Timestamp:        uint32(uint32(time.Date(2017, time.February, 23, 0, 0, 0, 0, time.UTC).Unix())),
		Height:           uint32(0),
		ConsensusData:    GenesisNonce,
		NextBookKeeper:   nextBookKeeper,
		Program: &program.Program{
			Code:      []byte{},
			Parameter: []byte{byte(vm.PUSHT)},
		},
	}
	//transaction
	trans := &tx.Transaction{
		TxType:         tx.BookKeeping,
		PayloadVersion: payload.BookKeepingPayloadVersion,
		Payload: &payload.BookKeeping{
			Nonce: GenesisNonce,
		},
		Attributes:    []*tx.TxAttribute{},
		UTXOInputs:    []*utxo.UTXOTxInput{},
		BalanceInputs: []*tx.BalanceTxInput{},
		Outputs:       []*utxo.TxOutput{},
		Programs:      []*program.Program{},
	}
	//block
	issue := tx.NewIssueToken(tx.ONTToken, tx.ONGToken)
	genesisBlock := &Block{
		Header: genesisHeader,
		Transactions: []*tx.Transaction{
			trans,
			tx.ONTToken,
			tx.ONGToken,
			issue,
		},
	}
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
	b.Header.TransactionsRoot = hash
	return nil

}

func (bd *Block) SerializeUnsigned(w io.Writer) error {
	return bd.Header.SerializeUnsigned(w)
}

func CalculateBouns(inputs []*utxo.UTXOTxInput, ignoreClaimed bool) (Fixed64, error) {
	unclaimed := make([]*utxo.SpentCoin, 0)
	group := make(map[Uint256][]uint16, 0)
	for _, v := range inputs {
		if _, ok := group[v.ReferTxID]; !ok {
			group[v.ReferTxID] = make([]uint16, 0)
		}
		group[v.ReferTxID] = append(group[v.ReferTxID], v.ReferTxOutputIndex)
	}
	for k, v := range group {
		claimable, err := DefaultLedger.Store.GetUnclaimed(k)
		if err != nil {
			if ignoreClaimed {
				continue
			} else {
				return 0, err
			}
		}
		for _, m := range v {
			if value, ok := claimable[m]; !ok {
				if ignoreClaimed {
					continue
				} else {
					return 0, err
				}
				unclaimed = append(unclaimed, value)
			}
			unclaimed = append(unclaimed, claimable[m])
		}
	}
	return CalculateBonusInternal(unclaimed)
}

func CalculateBonusInternal(unclaimed []*utxo.SpentCoin) (Fixed64, error) {
	gl := uint32(len(GenerationAmount))
	var amountClaimed uint64
	type Temp struct {
		StartHeight uint32
		EndHeight   uint32
	}
	group := make(map[Temp]Fixed64, 0)
	for _, v := range unclaimed {
		group[Temp{v.StartHeight, v.EndHeight}] = group[Temp{v.StartHeight, v.EndHeight}] + v.Output.Value
	}
	for k, v := range group {
		var amount uint32 = 0
		ustart := k.StartHeight / DecrementInterval
		if ustart < gl {
			istart := k.StartHeight % DecrementInterval
			uend := k.EndHeight / DecrementInterval
			iend := k.EndHeight % DecrementInterval
			if uend >= gl {
				uend = gl
				iend = 0
			}
			if iend == 0 {
				uend--
				iend = DecrementInterval
			}
			for {
				if ustart >= uend {
					break
				}
				amount += (DecrementInterval - istart) * GenerationAmount[ustart]
				ustart++
				istart = 0
			}
			amount += iend * GenerationAmount[ustart]
		}
		var (
			hash        Uint256
			err         error
			startSysFee Fixed64
			endSysFee   Fixed64
		)

		if k.StartHeight == 0 {
			startSysFee = Fixed64(0)
		} else {
			hash, err = DefaultLedger.Store.GetBlockHash(k.StartHeight - 1)
			if err != nil {
				return Fixed64(0), err
			}
			startSysFee, err = DefaultLedger.Store.GetSysFeeAmount(hash)
			if err != nil {
				return Fixed64(0), err
			}
		}

		hash, err = DefaultLedger.Store.GetBlockHash(k.EndHeight - 1)
		if err != nil {
			return Fixed64(0), err
		}

		endSysFee, err = DefaultLedger.Store.GetSysFeeAmount(hash)
		if err != nil {
			return Fixed64(0), err
		}
		amount += uint32(endSysFee.GetData()) - uint32(startSysFee.GetData())
		amountClaimed += uint64(v.GetData()) / uint64(tx.OntRegisterAmount) * uint64(amount)

	}
	return Fixed64(amountClaimed), nil
}
