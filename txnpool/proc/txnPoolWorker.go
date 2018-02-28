package proc

import (
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	tx "github.com/Ontology/core/types"
	"github.com/Ontology/eventbus/eventhub"
	tc "github.com/Ontology/txnpool/common"
	"sync"
	"time"
)

type pendingTxn struct {
	txn     *tx.Transaction // That is unverified or on the verifying process
	worker  *txnPoolWorker  // Which worker handles it
	valTime time.Time       // The start time
	req     *tc.VerifyReq   // Req cache
	flag    uint8           // For different types of verification
	retries uint8           // For resend to validator when time out before verified
	ret     []*tc.TXNAttr   // verified results
}

type txnPoolWorker struct {
	mu             sync.RWMutex
	workId         uint8                          // Worker ID
	rcvTXNCh       chan *tx.Transaction           // The channel of receive transaction
	rspCh          chan *tc.VerifyRsp             // The channel of verified response
	server         *TXNPoolServer                 // The txn pool server pointer
	timer          *time.Timer                    // The timer of reverifying
	stopCh         chan bool                      // stop routine
	pendingTxnList map[common.Uint256]*pendingTxn // The transaction on the verifying process
}

func (worker *txnPoolWorker) init(workID uint8, s *TXNPoolServer) {
	worker.rcvTXNCh = make(chan *tx.Transaction, tc.MAXPENDINGTXN)
	worker.pendingTxnList = make(map[common.Uint256]*pendingTxn)
	worker.rspCh = make(chan *tc.VerifyRsp, tc.MAXPENDINGTXN)
	worker.stopCh = make(chan bool)
	worker.workId = workID
	worker.server = s
}

func (worker *txnPoolWorker) GetTxnStatus(hash common.Uint256) *tc.TXNEntry {
	worker.mu.RLock()
	defer worker.mu.RUnlock()

	pt, ok := worker.pendingTxnList[hash]
	if !ok {
		return nil
	}

	txnEntry := &tc.TXNEntry{
		Txn:   pt.txn,
		Attrs: pt.ret,
		Fee:   pt.txn.GetTotalFee(),
	}
	return txnEntry
}

func (worker *txnPoolWorker) handleRsp(rsp *tc.VerifyRsp) {
	if rsp.WorkerId != worker.workId {
		return
	}

	worker.mu.Lock()
	defer worker.mu.Unlock()

	pt, ok := worker.pendingTxnList[rsp.TxnHash]
	if !ok {
		return
	}
	if rsp.Ok != true {
		//Verify fail
		log.Info(fmt.Sprintf("Validator %d: Transaction %x invalid",
			rsp.ValidatorID, rsp.TxnHash))
		delete(worker.pendingTxnList, rsp.TxnHash)
		worker.server.removePendingTxn(rsp.TxnHash)
		return
	}

	if pt.flag&(0x1<<(rsp.ValidatorID-1)) == 0 {
		retAttr := &tc.TXNAttr{
			Height:      rsp.Height,
			ValidatorID: rsp.ValidatorID,
			Ok:          rsp.Ok,
		}
		pt.flag |= (0x1 << (rsp.ValidatorID - 1))
		pt.ret = append(pt.ret, retAttr)
	}

	if pt.flag&0xf == tc.VERIFYMASK {
		worker.putTxnPool(pt)
		delete(worker.pendingTxnList, rsp.TxnHash)
	}
}

/* Check if the transaction need to be sent to validator to verify when time out
 * Todo: Going through the list will take time if the list is too long, need to
 * change the algorithm later
 */
func (worker *txnPoolWorker) handleTimeoutEvent() {
	if len(worker.pendingTxnList) <= 0 {
		return
	}

	/* Go through the pending list, for those unverified txns,
	 * resend them to the validators
	 */
	for k, v := range worker.pendingTxnList {
		if v.flag&0xf != tc.VERIFYMASK && time.Now().Sub(v.valTime) >=
			tc.EXPIREINTERVAL {
			if v.retries < tc.MAXRETRIES {
				worker.reVerifyTxn(k)
				v.retries++
			} else {
				// Todo: Retry exhausted, remove it from pendingTxnList
				worker.mu.Lock()
				delete(worker.pendingTxnList, k)
				worker.mu.Unlock()
			}
		}
	}
}

func (worker *txnPoolWorker) putTxnPool(pt *pendingTxn) bool {
	txnEntry := &tc.TXNEntry{
		Txn:   pt.txn,
		Attrs: pt.ret,
		Fee:   pt.txn.GetTotalFee(),
	}
	worker.server.AddTxnList(txnEntry)
	worker.server.removePendingTxn(pt.txn.Hash())
	return true
}

func (worker *txnPoolWorker) verifyTxn(txn *tx.Transaction) {
	if txn := worker.server.getTransaction(txn.Hash()); txn != nil {
		log.Info(fmt.Sprintf("Transaction %x already in the txn pool",
			txn.Hash()))
		return
	}

	if _, ok := worker.pendingTxnList[txn.Hash()]; ok {
		log.Info(fmt.Sprintf("Transaction %x already in the verifying process",
			txn.Hash()))
		return
	}
	// Construct the request and send it to each validator server to verify
	req := &tc.VerifyReq{
		WorkerId: worker.workId,
		Txn:      txn,
	}

	worker.sendReq2Validator(req)

	// Construct the pending transaction
	pt := &pendingTxn{
		txn:     txn,
		worker:  worker,
		req:     req,
		flag:    0,
		retries: 0,
	}
	// Add it to the pending transaction list
	worker.mu.Lock()
	worker.pendingTxnList[txn.Hash()] = pt
	worker.mu.Unlock()
	// Record the time per a txn
	pt.valTime = time.Now()
}

func (worker *txnPoolWorker) reVerifyTxn(txnHash common.Uint256) {
	// Todo: add retry logic
	pt, ok := worker.pendingTxnList[txnHash]
	if !ok {
		return
	}

	if pt.flag&0xf != tc.VERIFYMASK {
		worker.sendReq2Validator(pt.req)
	}

	// Update the verifying time
	pt.valTime = time.Now()
}

func (worker *txnPoolWorker) sendReq2Validator(req *tc.VerifyReq) (send bool) {
	defer func() {
		if recover() != nil {
			send = false
		}
	}()

	pid := worker.server.GetPID(tc.VerifyRspActor)
	if pid == nil {
		log.Info("VerifyRspActor not exist")
		return false
	}

	event := eventhub.Event{Publisher: pid, Message: req, Topic: tc.TOPIC,
		Policy: eventhub.PUBLISH_POLICY_ALL}
	worker.server.publishEvent(&event)
	return true
}

func (worker *txnPoolWorker) start() {
	worker.timer = time.NewTimer(time.Second * tc.EXPIREINTERVAL)
	for {
		select {
		case <-worker.stopCh:
			worker.server.wg.Done()
			return
		case rcvTxn, ok := <-worker.rcvTXNCh:
			if ok {
				// Verify rcvTxn
				worker.verifyTxn(rcvTxn)
			}
		case <-worker.timer.C:
			worker.handleTimeoutEvent()
			worker.timer.Stop()
			worker.timer.Reset(time.Second * tc.EXPIREINTERVAL)
		case rsp, ok := <-worker.rspCh:
			if ok {
				/* Handle the response from validator, if all of cases
				 * are verified, put it to txnPool
				 */
				worker.handleRsp(rsp)
			}
		}
	}
}

func (worker *txnPoolWorker) stop() {
	if worker.timer != nil {
		worker.timer.Stop()
	}
	if worker.rcvTXNCh != nil {
		close(worker.rcvTXNCh)
	}
	if worker.rspCh != nil {
		close(worker.rspCh)
	}

	if worker.stopCh != nil {
		worker.stopCh <- true
		close(worker.stopCh)
	}
}
