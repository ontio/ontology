/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package types

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
)

type TxData struct {
	Id   int
	Data []byte
}

type Tx struct {
	Id int
	Tx *Transaction
}

func parallelDeserializeTxs(r io.Reader, n int) ([]*Transaction, []common.Uint256, error) {
	K := 8
	inC := make(chan *TxData, 1024)
	outC := make(chan *Tx, 1024)
	errC := make(chan error, K) // all deserializer worker can report one error
	quitC := make(chan struct{})

	// start K deserializer routiner
	deserializerWg := sync.WaitGroup{}
	for i := 0; i < K; i++ {
		// worker
		go func(txdataC <-chan *TxData, txC chan<- *Tx, errC chan<- error) {
			deserializerWg.Add(1)
			defer deserializerWg.Done()

			for {
				select {
				case data := <-txdataC:
					if data == nil {
						// input channel closed, no more input
						return
					}
					transaction := new(Transaction)
					err := transaction.Deserialize(bytes.NewBuffer(data.Data))
					if err != nil {
						// report err with errC
						errC <- err
						return
					}
					// send deserialized tx to outC
					txC <- &Tx{
						Id: data.Id,
						Tx: transaction,
					}
				case <-quitC:
					// got error, quit
					return
				}
			}
		}(inC, outC, errC)
	}

	// tx summarizer
	summarizerWg := sync.WaitGroup{}
	txmap := make(map[int]*Transaction)
	go func(txC <-chan *Tx) {
		summarizerWg.Add(1)
		defer summarizerWg.Done()

		for {
			select {
			case tx := <-txC:
				if tx == nil {
					// no more deserialized tx
					return
				}
				txmap[tx.Id] = tx.Tx
			case <-quitC:
				// got error, quit
				return
			}
		}
	}(outC)

	// tx buffer producer
	for i := 0; i < n; i++ {
		txData, err := serialization.ReadVarBytes(r)
		if err != nil {
			close(quitC)
			return nil, nil, err
		}
		// check if any error pending
		select {
		case err := <-errC:
			close(quitC)
			return nil, nil, err
		default:
		}
		inC <- &TxData{
			Id:   i,
			Data: txData,
		}
	}
	// all tx buffer send to inC
	close(inC)
	// wait deserializer workers done
	deserializerWg.Wait()
	// all tx deserializer done, all deserialized tx have sent to outC
	close(outC)
	// wait summarizer done
	summarizerWg.Wait()

	// this is not necessary, just in case some routine not exit
	close(quitC)

	// reorder all tx
	txs := make([]*Transaction, 0, n)
	hashes := make([]common.Uint256, 0, n)
	for i := 0; i < n; i++ {
		tx := txmap[i]
		if tx == nil {
			return nil, nil, fmt.Errorf("no tx for %d", i)
		}
		txs = append(txs, tx)
		hashes = append(hashes, tx.Hash())
	}
	return txs, hashes, nil
}
