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
	Pid       *actor.PID       // The actor pid of the registered validator
	CheckType types.VerifyType // The validator's type: stateless/stateful
}

type serverPendingTx struct {
	tx     *tx.Transaction // Pending tx
	sender *actor.PID      // Indicate which sender tx is from
}

type pendingBlock struct {
	mu             sync.RWMutex
	sender         *actor.PID                            // Consensus PID
	height         uint32                                // The block height
	processedTxs   map[common.Uint256]*tc.VerifyTxResult // Transaction which has been processed
	unProcessedTxs map[common.Uint256]*tx.Transaction    // Transaction which is not processed
	stopCh         chan bool                             // Sync call, right now, server only can handle one by one
}

type TXPoolServer struct {
	mu            sync.RWMutex                        // Sync mutex
	wg            sync.WaitGroup                      // Worker sync
	workers       []txPoolWorker                      // Worker pool
	workersNum    uint8                               // The number of concurrent workers
	txPool        *tc.TXPool                          // The tx pool that holds the valid transaction
	allPendingTxs map[common.Uint256]*serverPendingTx // The txs that server is processing
	pendingBlock  *pendingBlock                       // The block that server is processing
	actors        map[tc.ActorType]*actor.PID         // The actors running in the server
	validators    map[string]Validator                // The registered validators
	stats         txStats                             // The transaction statstics
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
	s.allPendingTxs = make(map[common.Uint256]*serverPendingTx)
	s.actors = make(map[tc.ActorType]*actor.PID)
	s.validators = make(map[string]Validator)
	s.stats = txStats{count: make([]uint64, tc.MAXSTATS-1)}

	s.pendingBlock = &pendingBlock{
		processedTxs:   make(map[common.Uint256]*tc.VerifyTxResult, 0),
		unProcessedTxs: make(map[common.Uint256]*tx.Transaction, 0),
		stopCh:         make(chan bool),
	}

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

func (s *TXPoolServer) sendRsp2Client(sender *actor.PID, hash common.Uint256,
	err errors.ErrCode) {
	res := &tc.TxRsp{
		Hash:    hash,
		ErrCode: err,
	}
	sender.Request(res, s.GetPID(tc.TxActor))
}

func (s *TXPoolServer) checkPendingBlockOk(hash common.Uint256, err errors.ErrCode) {
	// Check if the tx is in pending block, if yes, move it to
	// the verified tx list
	s.pendingBlock.mu.Lock()
	defer s.pendingBlock.mu.Unlock()

	tx, ok := s.pendingBlock.unProcessedTxs[hash]
	if !ok {
		return
	}

	// Todo:
	entry := &tc.VerifyTxResult{
		Height:  s.pendingBlock.height,
		Tx:      tx,
		ErrCode: err,
	}

	s.pendingBlock.processedTxs[hash] = entry
	delete(s.pendingBlock.unProcessedTxs, hash)

	// Check if the block has been verified, if yes, send rsp to
	// the actor bus
	if len(s.pendingBlock.unProcessedTxs) == 0 {
		rsp := &tc.VerifyBlockRsp{
			TxnPool: make([]*tc.VerifyTxResult, len(s.pendingBlock.processedTxs)),
		}
		for _, v := range s.pendingBlock.processedTxs {
			rsp.TxnPool = append(rsp.TxnPool, v)
		}

		if s.pendingBlock.sender != nil {
			s.pendingBlock.sender.Tell(rsp)
		}

		if s.pendingBlock.stopCh != nil {
			s.pendingBlock.stopCh <- true
		}
	}
}

func (s *TXPoolServer) removePendingTx(hash common.Uint256, err errors.ErrCode) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pt, ok := s.allPendingTxs[hash]
	if !ok {
		return
	}
	if pt.sender != nil {
		s.sendRsp2Client(pt.sender, hash, err)
	}

	delete(s.allPendingTxs, hash)

	// Check if the tx is in the pending block and the pending block is verified
	s.checkPendingBlockOk(hash, err)
}

func (s *TXPoolServer) setPendingTx(tx *tx.Transaction, sender *actor.PID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ok := s.allPendingTxs[tx.Hash()]; ok != nil {
		log.Info("Transaction already in the verifying process",
			tx.Hash())
		return false
	}

	pt := &serverPendingTx{
		tx:     tx,
		sender: sender,
	}

	s.allPendingTxs[tx.Hash()] = pt
	return true
}

func (s *TXPoolServer) assginTXN2Worker(tx *tx.Transaction, sender *actor.PID) (
	assign bool) {
	defer func() {
		if recover() != nil {
			assign = false
		}
	}()

	if tx == nil {
		return
	}

	if ok := s.setPendingTx(tx, sender); !ok {
		s.increaseStats(tc.DuplicateStats)
		if sender != nil {
			s.sendRsp2Client(sender, tx.Hash(), errors.ErrDuplicatedTx)
		}
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

func (s *TXPoolServer) getValidatorPID(id string) *actor.PID {
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

func (s *TXPoolServer) getTxPool(byCount bool) []*tc.TXEntry {
	return s.txPool.GetTxPool(byCount)
}

func (s *TXPoolServer) getPendingTxs(byCount bool) []*tx.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ret := make([]*tx.Transaction, 0, len(s.allPendingTxs))
	for _, v := range s.allPendingTxs {
		ret = append(ret, v.tx)
	}
	return ret
}

func (s *TXPoolServer) cleanTransactionList(txs []*tx.Transaction) error {
	return s.txPool.CleanTransactionList(txs)
}

func (s *TXPoolServer) addTxList(txEntry *tc.TXEntry) bool {
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

func (s *TXPoolServer) checkTx(hash common.Uint256) bool {
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

func (s *TXPoolServer) getTxStatusReq(hash common.Uint256) *tc.TxStatus {
	for i := uint8(0); i < s.workersNum; i++ {
		ret := s.workers[i].GetTxStatus(hash)
		if ret != nil {
			return ret
		}
	}

	return s.txPool.GetTxStatus(hash)
}

func (s *TXPoolServer) getTransactionCount() int {
	return s.txPool.GetTransactionCount()
}

func (s *TXPoolServer) verifyBlock(req *tc.VerifyBlockReq, sender *actor.PID) {
	if req == nil || len(req.Txs) == 0 {
		return
	}
	s.pendingBlock.sender = sender
	s.pendingBlock.height = req.Height

	for _, tx := range req.Txs {
		/* Check if the tx is in the tx pool, if not, send it to
		 * valdiator to verify and add it to pending block list
		 */
		ret := s.txPool.GetTxStatus(tx.Hash())
		if ret == nil {
			s.assginTXN2Worker(tx, nil)
			s.pendingBlock.unProcessedTxs[tx.Hash()] = tx
			continue
		}

		/* Check the verified height >= the block height, if yes,
		 * add it to the response list.
		 */
		ok := false
		for _, v := range ret.Attrs {
			if v.Type == types.Statefull &&
				v.Height >= req.Height {
				entry := &tc.VerifyTxResult{
					Tx:      tx,
					Height:  v.Height,
					ErrCode: v.ErrCode,
				}
				s.pendingBlock.processedTxs[tx.Hash()] = entry
				ok = true
				break
			}
		}

		// Re-verify it
		if !ok {
			s.assginTXN2Worker(tx, nil)
			s.pendingBlock.unProcessedTxs[tx.Hash()] = tx
		}
	}
	<-s.pendingBlock.stopCh
}
