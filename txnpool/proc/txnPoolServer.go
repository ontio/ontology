package proc

import (
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	tx "github.com/Ontology/core/types"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/eventbus/eventhub"
	tc "github.com/Ontology/txnpool/common"
	"sort"
	"sync"
)

type txnStats struct {
	sync.RWMutex
	count []uint64
}

type TXNPoolServer struct {
	mu             sync.RWMutex
	wg             sync.WaitGroup
	workers        []txnPoolWorker
	workersNum     uint8
	txnPool        *tc.TXNPool
	allPendingTxns map[common.Uint256]*tx.Transaction
	eh             *eventhub.EventHub
	actors         map[tc.ActorType]*actor.PID
	stats          txnStats
}

func NewTxnPoolServer(num uint8) *TXNPoolServer {
	s := &TXNPoolServer{}
	s.init(num)
	return s
}

func (s *TXNPoolServer) init(num uint8) {
	// Initial txnPool
	s.txnPool = &tc.TXNPool{}
	s.txnPool.Init()
	s.allPendingTxns = make(map[common.Uint256]*tx.Transaction)
	s.actors = make(map[tc.ActorType]*actor.PID)
	s.stats = txnStats{count: make([]uint64, tc.MAXSTATS-1)}

	// Create the given concurrent workers
	s.workers = make([]txnPoolWorker, num)
	s.workersNum = num
	// Initial and start the workers
	for i := uint8(0); i < num; i++ {
		s.wg.Add(1)
		s.workers[i].init(i, s)
		go s.workers[i].start()
	}
}

func (s *TXNPoolServer) removePendingTxn(hash common.Uint256) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.allPendingTxns, hash)
}

func (s *TXNPoolServer) setPendingTxn(txn *tx.Transaction) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ok := s.allPendingTxns[txn.Hash()]; ok != nil {
		log.Info("Transaction %x already in the verifying process",
			txn.Hash())
		return false
	}

	s.allPendingTxns[txn.Hash()] = txn
	return true
}

func (s *TXNPoolServer) assginTXN2Worker(txn *tx.Transaction) (
	assign bool) {
	defer func() {
		if recover() != nil {
			assign = false
		}
	}()

	if txn == nil {
		return
	}

	if ok := s.setPendingTxn(txn); !ok {
		s.increaseStats(tc.DuplicateStats)
		return false
	}
	// Add the rcvTxn to the worker
	lb := make(tc.LBSlice, s.workersNum)
	for i := uint8(0); i < s.workersNum; i++ {
		entry := tc.LB{Size: len(s.workers[i].pendingTxnList),
			WorkerID: i,
		}
		lb[i] = entry
	}
	sort.Sort(lb)
	s.workers[lb[0].WorkerID].rcvTXNCh <- txn

	return true
}

func (s *TXNPoolServer) assignRsp2Worker(rsp *tc.VerifyRsp) (assign bool) {
	defer func() {
		if recover() != nil {
			assign = false
		}
	}()

	if rsp == nil {
		return
	}

	if rsp.WorkerId >= 0 && rsp.WorkerId < s.workersNum {
		s.workers[rsp.WorkerId].rspCh <- rsp
	}

	if rsp.Ok {
		s.increaseStats(tc.SuccessStats)
	} else {
		s.increaseStats(tc.FailureStats)
		if rsp.ValidatorID == uint8(tc.SignatureV) {
			s.increaseStats(tc.SigErrStats)
		} else {
			s.increaseStats(tc.StateErrStats)
		}
	}
	return true
}

func (s *TXNPoolServer) GetPID(actor tc.ActorType) *actor.PID {
	if actor < tc.TxActor || actor >= tc.MAXACTOR {
		return nil
	}

	return s.actors[actor]
}

func (s *TXNPoolServer) RegisterActor(actor tc.ActorType, pid *actor.PID) {
	s.actors[actor] = pid
}

func (s *TXNPoolServer) UnRegisterActor(actor tc.ActorType) {
	delete(s.actors, actor)
}

func (s *TXNPoolServer) Stop() {
	for _, v := range s.actors {
		v.Stop()
	}
	//Stop worker
	for i := uint8(0); i < s.workersNum; i++ {
		s.workers[i].stop()
	}
	s.wg.Wait()
}

func (s *TXNPoolServer) SetEventHub(eh *eventhub.EventHub) {
	s.eh = eh
}

func (s *TXNPoolServer) GetEventHub() *eventhub.EventHub {
	return s.eh
}

func (s *TXNPoolServer) publishEvent(event *eventhub.Event) {
	s.eh.Publish(event)
}

func (s *TXNPoolServer) getTransaction(hash common.Uint256) *tx.Transaction {
	return s.txnPool.GetTransaction(hash)
}

func (s *TXNPoolServer) GetTxnPool(byCount bool) []*tc.TXNEntry {
	return s.txnPool.GetTxnPool(byCount)
}

func (s *TXNPoolServer) GetPendingTxs(byCount bool) []*tx.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ret := make([]*tx.Transaction, 0, len(s.allPendingTxns))
	for _, v := range s.allPendingTxns {
		ret = append(ret, v)
	}
	return ret
}

func (s *TXNPoolServer) GetUnverifiedTxs(txs []*tx.Transaction) []*tx.Transaction {
	if len(txs) == 0 {
		return nil
	}
	return s.txnPool.GetUnverifiedTxs(txs)
}

func (s *TXNPoolServer) CleanTransactionList(txns []*tx.Transaction) error {
	return s.txnPool.CleanTransactionList(txns)
}

func (s *TXNPoolServer) AddTxnList(txnEntry *tc.TXNEntry) bool {
	ret := s.txnPool.AddTxnList(txnEntry)
	if !ret {
		s.increaseStats(tc.DuplicateStats)
	}
	return ret
}

func (s *TXNPoolServer) increaseStats(v tc.TxnStatsType) {
	s.stats.Lock()
	defer s.stats.Unlock()
	s.stats.count[v-1]++
}

func (s *TXNPoolServer) getStats() *[]uint64 {
	s.stats.RLock()
	defer s.stats.RUnlock()
	ret := make([]uint64, 0, len(s.stats.count))
	for _, v := range s.stats.count {
		ret = append(ret, v)
	}
	return &ret
}

func (s *TXNPoolServer) CheckTxn(hash common.Uint256) bool {
	// Check if the tx is in pending list
	s.mu.RLock()
	if ok := s.allPendingTxns[hash]; ok != nil {
		s.mu.RUnlock()
		return true
	}
	s.mu.RUnlock()

	// Check if the tx is in txn pool
	if res := s.txnPool.GetTransaction(hash); res != nil {
		return true
	}

	return false
}

func (s *TXNPoolServer) GetTxnStatusReq(hash common.Uint256) *tc.TXNEntry {
	for i := uint8(0); i < s.workersNum; i++ {
		ret := s.workers[i].GetTxnStatus(hash)
		if ret != nil {
			return ret
		}
	}

	return s.txnPool.GetTxnStatus(hash)
}
