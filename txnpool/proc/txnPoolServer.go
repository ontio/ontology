package proc

import (
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	tx "github.com/Ontology/core/types"
	"github.com/Ontology/errors"
	"github.com/Ontology/eventbus/actor"
	tc "github.com/Ontology/txnpool/common"
	"github.com/Ontology/validator/types"
	"sort"
	"sync"
)

type txStats struct {
	sync.RWMutex
	count []uint64
}

type Validator struct {
	Pid       *actor.PID
	CheckType types.VerifyType
}

type TXPoolServer struct {
	mu            sync.RWMutex
	wg            sync.WaitGroup
	workers       []txPoolWorker
	workersNum    uint8
	txPool        *tc.TXPool
	allPendingTxs map[common.Uint256]*tx.Transaction
	actors        map[tc.ActorType]*actor.PID
	validators    map[string]Validator
	stats         txStats
}

func NewTxPoolServer(num uint8) *TXPoolServer {
	s := &TXPoolServer{}
	s.init(num)
	return s
}

func (s *TXPoolServer) init(num uint8) {
	// Initial txnPool
	s.txPool = &tc.TXPool{}
	s.txPool.Init()
	s.allPendingTxs = make(map[common.Uint256]*tx.Transaction)
	s.actors = make(map[tc.ActorType]*actor.PID)
	s.validators = make(map[string]Validator)
	s.stats = txStats{count: make([]uint64, tc.MAXSTATS-1)}

	// Create the given concurrent workers
	s.workers = make([]txPoolWorker, num)
	s.workersNum = num
	// Initial and start the workers
	for i := uint8(0); i < num; i++ {
		s.wg.Add(1)
		s.workers[i].init(i, s)
		go s.workers[i].start()
	}
}

func (s *TXPoolServer) removePendingTx(hash common.Uint256) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.allPendingTxs, hash)
}

func (s *TXPoolServer) setPendingTx(tx *tx.Transaction) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ok := s.allPendingTxs[tx.Hash()]; ok != nil {
		log.Info("Transaction already in the verifying process",
			tx.Hash())
		return false
	}

	s.allPendingTxs[tx.Hash()] = tx
	return true
}

func (s *TXPoolServer) assginTXN2Worker(tx *tx.Transaction) (
	assign bool) {
	defer func() {
		if recover() != nil {
			assign = false
		}
	}()

	if tx == nil {
		return
	}

	if ok := s.setPendingTx(tx); !ok {
		s.increaseStats(tc.DuplicateStats)
		return false
	}
	// Add the rcvTxn to the worker
	lb := make(tc.LBSlice, s.workersNum)
	for i := uint8(0); i < s.workersNum; i++ {
		entry := tc.LB{Size: len(s.workers[i].pendingTxList),
			WorkerID: i,
		}
		lb[i] = entry
	}
	sort.Sort(lb)
	s.workers[lb[0].WorkerID].rcvTXCh <- tx

	return true
}

func (s *TXPoolServer) assignRsp2Worker(rsp *types.CheckResponse) (assign bool) {
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

	if rsp.ErrCode == errors.ErrNoError {
		s.increaseStats(tc.SuccessStats)
	} else {
		s.increaseStats(tc.FailureStats)
		if rsp.Type == types.Stateless {
			s.increaseStats(tc.SigErrStats)
		} else {
			s.increaseStats(tc.StateErrStats)
		}
	}
	return true
}

func (s *TXPoolServer) GetPID(actor tc.ActorType) *actor.PID {
	if actor < tc.TxActor || actor >= tc.MAXACTOR {
		return nil
	}

	return s.actors[actor]
}

func (s *TXPoolServer) RegisterActor(actor tc.ActorType, pid *actor.PID) {
	s.actors[actor] = pid
}

func (s *TXPoolServer) UnRegisterActor(actor tc.ActorType) {
	delete(s.actors, actor)
}

func (s *TXPoolServer) registerValidator(id string, v Validator) {
	s.validators[id] = v
}

func (s *TXPoolServer) unRegisterValidator(id string) {
	delete(s.validators, id)
}

func (s *TXPoolServer) GetValidatorPID(id string) *actor.PID {
	v, ok := s.validators[id]
	if !ok {
		return nil
	}
	return v.Pid
}

func (s *TXPoolServer) Stop() {
	for _, v := range s.actors {
		v.Stop()
	}
	//Stop worker
	for i := uint8(0); i < s.workersNum; i++ {
		s.workers[i].stop()
	}
	s.wg.Wait()
}

func (s *TXPoolServer) getTransaction(hash common.Uint256) *tx.Transaction {
	return s.txPool.GetTransaction(hash)
}

func (s *TXPoolServer) GetTxPool(byCount bool) []*tc.TXEntry {
	return s.txPool.GetTxPool(byCount)
}

func (s *TXPoolServer) GetPendingTxs(byCount bool) []*tx.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ret := make([]*tx.Transaction, 0, len(s.allPendingTxs))
	for _, v := range s.allPendingTxs {
		ret = append(ret, v)
	}
	return ret
}

func (s *TXPoolServer) GetUnverifiedTxs(txs []*tx.Transaction) []*tx.Transaction {
	if len(txs) == 0 {
		return nil
	}
	return s.txPool.GetUnverifiedTxs(txs)
}

func (s *TXPoolServer) CleanTransactionList(txs []*tx.Transaction) error {
	return s.txPool.CleanTransactionList(txs)
}

func (s *TXPoolServer) AddTxList(txEntry *tc.TXEntry) bool {
	ret := s.txPool.AddTxList(txEntry)
	if !ret {
		s.increaseStats(tc.DuplicateStats)
	}
	return ret
}

func (s *TXPoolServer) increaseStats(v tc.TxnStatsType) {
	s.stats.Lock()
	defer s.stats.Unlock()
	s.stats.count[v-1]++
}

func (s *TXPoolServer) getStats() *[]uint64 {
	s.stats.RLock()
	defer s.stats.RUnlock()
	ret := make([]uint64, 0, len(s.stats.count))
	for _, v := range s.stats.count {
		ret = append(ret, v)
	}
	return &ret
}

func (s *TXPoolServer) CheckTx(hash common.Uint256) bool {
	// Check if the tx is in pending list
	s.mu.RLock()
	if ok := s.allPendingTxs[hash]; ok != nil {
		s.mu.RUnlock()
		return true
	}
	s.mu.RUnlock()

	// Check if the tx is in txn pool
	if res := s.txPool.GetTransaction(hash); res != nil {
		return true
	}

	return false
}

func (s *TXPoolServer) GetTxStatusReq(hash common.Uint256) *tc.TxStatus {
	for i := uint8(0); i < s.workersNum; i++ {
		ret := s.workers[i].GetTxStatus(hash)
		if ret != nil {
			return ret
		}
	}

	return s.txPool.GetTxStatus(hash)
}
