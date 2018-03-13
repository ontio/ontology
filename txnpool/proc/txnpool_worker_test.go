package proc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/core/types"
	"github.com/Ontology/errors"
	tc "github.com/Ontology/txnpool/common"
	vt "github.com/Ontology/validator/types"
	"testing"
	"time"
)

func TestWorker(t *testing.T) {
	fmt.Println("Starting worker test")
	s := NewTxPoolServer(tc.MAXWORKERNUM)
	if s == nil {
		t.Error("Test case: new tx pool server failed")
		return
	}

	worker := &txPoolWorker{}
	worker.init(tc.MAXWORKERNUM, s)

	s.wg.Add(1)

	go worker.start()

	/* Case 1: For the given tx, validators reply
	 * with ErrNoError code. And tx pool should
	 * return the tx with the specified hash
	 */
	worker.rcvTXCh <- txn
	time.Sleep(1 * time.Second)

	statelessRsp := &vt.CheckResponse{
		WorkerId: tc.MAXWORKERNUM,
		ErrCode:  errors.ErrNoError,
		Hash:     txn.Hash(),
		Type:     vt.Stateless,
		Height:   0,
	}

	statefulRsp := &vt.CheckResponse{
		WorkerId: tc.MAXWORKERNUM,
		ErrCode:  errors.ErrNoError,
		Hash:     txn.Hash(),
		Type:     vt.Statefull,
		Height:   0,
	}
	worker.rspCh <- statelessRsp
	worker.rspCh <- statefulRsp

	time.Sleep(1 * time.Second)
	ret := worker.server.getTransaction(txn.Hash())
	fmt.Println(ret)

	/* Case 2: Duplicate input tx, worker should reject
	 * it with the log
	 */
	worker.rcvTXCh <- txn

	/* Case 3: For the given tx, validators reply with
	 * Error code, worker should remove the tx from the
	 * pending list with the log
	 */
	time.Sleep(1 * time.Second)
	worker.server.cleanTransactionList([]*types.Transaction{txn})

	worker.rcvTXCh <- txn
	time.Sleep(1 * time.Second)

	statelessRsp = &vt.CheckResponse{
		WorkerId: tc.MAXWORKERNUM,
		ErrCode:  errors.ErrUnknown,
		Hash:     txn.Hash(),
		Type:     vt.Stateless,
		Height:   0,
	}

	statefulRsp = &vt.CheckResponse{
		WorkerId: tc.MAXWORKERNUM,
		ErrCode:  errors.ErrUnknown,
		Hash:     txn.Hash(),
		Type:     vt.Statefull,
		Height:   0,
	}
	worker.rspCh <- statelessRsp
	worker.rspCh <- statefulRsp

	/* Case 4: valdiators reply with invalid tx hash or invalid work id,
	 * worker should reject it
	 */
	time.Sleep(2 * time.Second)
	statelessRsp = &vt.CheckResponse{
		WorkerId: tc.MAXWORKERNUM,
		ErrCode:  errors.ErrUnknown,
		Hash:     txn.Hash(),
		Type:     vt.Stateless,
		Height:   0,
	}

	statefulRsp = &vt.CheckResponse{
		WorkerId: tc.MAXWORKERNUM + 1,
		ErrCode:  errors.ErrUnknown,
		Hash:     txn.Hash(),
		Type:     vt.Statefull,
		Height:   0,
	}
	worker.rspCh <- statelessRsp
	worker.rspCh <- statefulRsp

	/* Case 5: For the given tx, response time out, worker should
	 * retry verifying it till retries exhausted, and then remove
	 * it from the pending list
	 */
	time.Sleep(2 * time.Second)
	worker.rcvTXCh <- txn

	time.Sleep(10 * time.Second)

	/* Case 6: For the given tx, worker handle it once, if
	 * duplicate input the tx, worker should reject it with
	 * the log.
	 */
	worker.rcvTXCh <- txn
	worker.rcvTXCh <- txn

	/* Case 7: For the pending tx, worker should get the entry
	 * with the valid hash
	 */
	time.Sleep(1 * time.Second)
	txStatus := worker.GetTxStatus(txn.Hash())
	fmt.Println(txStatus)
	/* Case 8: Given the invalid hash, worker should return nil
	 */
	tempStr := "3369930accc1ddd067245e8edadcd9bea207ba5e1753ac18a51df77a343bfe83"
	hex, _ := hex.DecodeString(tempStr)
	var hash common.Uint256
	hash.Deserialize(bytes.NewReader(hex))
	txStatus = worker.GetTxStatus(hash)
	fmt.Println(txStatus)

	worker.stop()
	s.Stop()
	fmt.Println("Ending worker test")
}
