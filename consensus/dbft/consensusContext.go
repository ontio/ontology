package dbft

import (
	"fmt"
	"github.com/Ontology/account"
	. "github.com/Ontology/common"
	"github.com/Ontology/common/log"
	ser "github.com/Ontology/common/serialization"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/types"
	"github.com/Ontology/core/vote"
	"github.com/Ontology/crypto"
	msg "github.com/Ontology/net/message"
)

const ContextVersion uint32 = 0

type ConsensusContext struct {
	State           ConsensusState
	PrevHash        Uint256
	Height          uint32
	ViewNumber      byte
	BookKeepers     []*crypto.PubKey
	NextBookKeepers []*crypto.PubKey
	Owner           *crypto.PubKey
	BookKeeperIndex int
	PrimaryIndex    uint32
	Timestamp       uint32
	Nonce           uint64
	NextBookKeeper  Uint160
	Transactions    []*types.Transaction
	Signatures      [][]byte
	ExpectedView    []byte

	header *types.Block

	isBookKeeperChanged bool
	nmChangedblkHeight  uint32
}

func (cxt *ConsensusContext) M() int {
	log.Debug()
	return len(cxt.BookKeepers) - (len(cxt.BookKeepers)-1)/3
}

func NewConsensusContext() *ConsensusContext {
	log.Debug()
	return &ConsensusContext{}
}

func (cxt *ConsensusContext) ChangeView(viewNum byte) {
	log.Debug()
	p := (cxt.Height - uint32(viewNum)) % uint32(len(cxt.BookKeepers))
	cxt.State &= SignatureSent
	cxt.ViewNumber = viewNum
	if p >= 0 {
		cxt.PrimaryIndex = uint32(p)
	} else {
		cxt.PrimaryIndex = uint32(p) + uint32(len(cxt.BookKeepers))
	}

	if cxt.State == Initial {
		cxt.Transactions = nil
		cxt.Signatures = make([][]byte, len(cxt.BookKeepers))
		cxt.header = nil
	}
}

func (cxt *ConsensusContext) MakeChangeView() *msg.ConsensusPayload {
	log.Debug()
	cv := &ChangeView{
		NewViewNumber: cxt.ExpectedView[cxt.BookKeeperIndex],
	}
	cv.msgData.Type = ChangeViewMsg
	return cxt.MakePayload(cv)
}

func (cxt *ConsensusContext) MakeHeader() *types.Block {
	log.Debug()
	if cxt.Transactions == nil {
		return nil
	}
	if cxt.header == nil {
		txHash := []Uint256{}
		for _, t := range cxt.Transactions {
			txHash = append(txHash, t.Hash())
		}
		txRoot, err := crypto.ComputeRoot(txHash)
		if err != nil {
			return nil
		}
		blockRoot := ledger.DefLedger.GetBlockRootWithNewTxRoot(txRoot)
		stateRoot, err := ledger.DefLedger.GetCurrentStateRoot()
		if err != nil {
			return nil
		}
		header := &types.Header{
			Version:          ContextVersion,
			PrevBlockHash:    cxt.PrevHash,
			TransactionsRoot: txRoot,
			BlockRoot:        blockRoot,
			StateRoot:        stateRoot,
			Timestamp:        cxt.Timestamp,
			Height:           cxt.Height,
			ConsensusData:    cxt.Nonce,
			NextBookKeeper:   cxt.NextBookKeeper,
		}
		cxt.header = &types.Block{
			Header:       header,
			Transactions: []*types.Transaction{},
		}
	}
	return cxt.header
}

func (cxt *ConsensusContext) MakePayload(message ConsensusMessage) *msg.ConsensusPayload {
	log.Debug()
	message.ConsensusMessageData().ViewNumber = cxt.ViewNumber
	return &msg.ConsensusPayload{
		Version:         ContextVersion,
		PrevHash:        cxt.PrevHash,
		Height:          cxt.Height,
		BookKeeperIndex: uint16(cxt.BookKeeperIndex),
		Timestamp:       cxt.Timestamp,
		Data:            ser.ToArray(message),
		Owner:           cxt.Owner,
	}
}

func (cxt *ConsensusContext) MakePrepareRequest() *msg.ConsensusPayload {
	log.Debug()
	preReq := &PrepareRequest{
		Nonce:          cxt.Nonce,
		NextBookKeeper: cxt.NextBookKeeper,
		Transactions:   cxt.Transactions,
		Signature:      cxt.Signatures[cxt.BookKeeperIndex],
	}
	preReq.msgData.Type = PrepareRequestMsg
	return cxt.MakePayload(preReq)
}

func (cxt *ConsensusContext) MakePrepareResponse(signature []byte) *msg.ConsensusPayload {
	log.Debug()
	preRes := &PrepareResponse{
		Signature: signature,
	}
	preRes.msgData.Type = PrepareResponseMsg
	return cxt.MakePayload(preRes)
}

func (cxt *ConsensusContext) MakeBlockSignatures(signatures []SignaturesData) *msg.ConsensusPayload {
	log.Debug()
	sigs := &BlockSignatures{
		Signatures: signatures,
	}
	sigs.msgData.Type = BlockSignaturesMsg
	return cxt.MakePayload(sigs)
}

func (cxt *ConsensusContext) GetSignaturesCount() (count int) {
	log.Debug()
	count = 0
	for _, sig := range cxt.Signatures {
		if sig != nil {
			count += 1
		}
	}
	return count
}

func (cxt *ConsensusContext) GetStateDetail() string {

	return fmt.Sprintf("Initial: %t, Primary: %t, Backup: %t, RequestSent: %t, RequestReceived: %t, SignatureSent: %t, BlockGenerated: %t, ",
		cxt.State.HasFlag(Initial),
		cxt.State.HasFlag(Primary),
		cxt.State.HasFlag(Backup),
		cxt.State.HasFlag(RequestSent),
		cxt.State.HasFlag(RequestReceived),
		cxt.State.HasFlag(SignatureSent),
		cxt.State.HasFlag(BlockGenerated))

}

func (cxt *ConsensusContext) Reset(bkAccount *account.Account) {
	preHash := ledger.DefLedger.GetCurrentBlockHash()
	height := ledger.DefLedger.GetCurrentBlockHeight()
	header := cxt.MakeHeader()

	if height != cxt.Height || header == nil || header.Hash() != preHash || len(cxt.NextBookKeepers) == 0 {
		log.Info("[ConsensusContext] Calculate BookKeepers from db")
		var err error
		cxt.BookKeepers, err = vote.GetValidators([]*types.Transaction{})
		if err != nil {
			log.Error("[ConsensusContext] GetNextBookKeeper failed", err)
		}
	} else {
		cxt.BookKeepers = cxt.NextBookKeepers
	}

	cxt.State = Initial
	cxt.PrevHash = preHash
	cxt.Height = height + 1
	cxt.ViewNumber = 0
	cxt.BookKeeperIndex = -1
	cxt.NextBookKeepers = nil
	bookKeeperLen := len(cxt.BookKeepers)
	cxt.PrimaryIndex = cxt.Height % uint32(bookKeeperLen)
	cxt.Transactions = nil
	cxt.header = nil
	cxt.Signatures = make([][]byte, bookKeeperLen)
	cxt.ExpectedView = make([]byte, bookKeeperLen)

	for i := 0; i < bookKeeperLen; i++ {
		if bkAccount.PublicKey.X.Cmp(cxt.BookKeepers[i].X) == 0 {
			cxt.BookKeeperIndex = i
			cxt.Owner = cxt.BookKeepers[i]
			break
		}
	}

}
