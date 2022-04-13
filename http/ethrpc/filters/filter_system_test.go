package filters

import (
	"fmt"
	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ontio/ontology/events"
	"github.com/ontio/ontology/events/message"
	"math/big"
	"testing"
	"time"
)

func TestNewEventSystem(t *testing.T) {
	events.Init()

	es := NewEventSystem(nil)
	fmt.Println("es:", es)

	header := &types.Header{Number: big.NewInt(1)}
	chainEvtMsg := &message.ChainEventMsg{
		ChainEvent: &core.ChainEvent{
			Block: types.NewBlockWithHeader(header),
			Hash:  common2.Hash{},
			Logs:  make([]*types.Log, 0),
		},
	}
	events.DefActorPublisher.Publish(message.TOPIC_CHAIN_EVENT, chainEvtMsg)
	time.Sleep(time.Second)

	scEvt := make(message.EthSmartCodeEvent, 0)
	scEvt = append(scEvt, &types.Log{})
	ethLog := &message.EthSmartCodeEventMsg{Event: scEvt}
	events.DefActorPublisher.Publish(message.TOPIC_ETH_SC_EVENT, ethLog)
	time.Sleep(time.Second)

	pendingTxEvt := &message.PendingTxMsg{Event: []*types.Transaction{&types.Transaction{}}}
	events.DefActorPublisher.Publish(message.TOPIC_PENDING_TX_EVENT, pendingTxEvt)
	time.Sleep(time.Second)
}
