package vbft

import (
	"fmt"
	"testing"
	"time"

	"github.com/Ontology/common/log"
)

func TestEventTimer_StartProposalTimer(t *testing.T) {
	log.Init(log.Stdout)
	log.Log.SetDebugLevel(1)

	s := &Server{
		log: log.Log,
	}

	timer := NewEventTimer(s)
	if timer == nil {
		t.FailNow()
	}

	receivedCnt := 0
	go func() {
		for {
			select {
			case evt := <-timer.C:
				fmt.Printf("timer received: evt: %d, blkNum: %d \n", evt.evtType, evt.blockNum)
				receivedCnt++
			case <-timer.stopC:
				break
			}
		}
	}()

	timer.StartProposalTimer(1)
	timer.StartEndorsingTimer(2)
	timer.StartCommitTimer(3)

	time.Sleep(time.Second)

	timer.Stop()
	if receivedCnt != 3 {
		t.FailNow()
	}
}

func TestEventTimer_CancelProposalTimer(t *testing.T) {
	log.Init(log.Stdout)
	log.Log.SetDebugLevel(1)

	s := &Server{
		log: log.Log,
	}

	timer := NewEventTimer(s)
	if timer == nil {
		t.FailNow()
	}

	receivedCnt := 0
	go func() {
		for {
			select {
			case evt := <-timer.C:
				fmt.Printf("timer received: evt: %d, blkNum: %d \n", evt.evtType, evt.blockNum)
				receivedCnt++
			case <-timer.stopC:
				break
			}
		}
	}()

	timer.StartTimer(1, time.Millisecond*500)
	timer.StartTimer(2, time.Millisecond*600)
	timer.StartTimer(3, time.Millisecond*700)

	timer.CancelTimer(2)

	time.Sleep(time.Second)

	timer.Stop()
	if receivedCnt != 2 {
		t.FailNow()
	}
}

func TestEventTimer_Stop(t *testing.T) {
	log.Init(log.Stdout)
	log.Log.SetDebugLevel(1)

	s := &Server{
		log: log.Log,
	}

	timer := NewEventTimer(s)
	if timer == nil {
		t.FailNow()
	}

	go func() {
		for {
			select {
			case <-timer.C:
				t.FailNow()
			case <-timer.stopC:
				break
			}
		}
	}()

	timer.StartTimer(1, time.Millisecond*500)
	timer.StartTimer(2, time.Millisecond*100)
	timer.StartProposalTimer(3)
	timer.StartEndorsingTimer(4)
	timer.StartCommitTimer(5)

	// all timer should be cleared by stop()
	timer.Stop()

	time.Sleep(time.Second)
}

func TestEventTimer_DoubleStartTimer(t *testing.T) {
	log.Init(log.Stdout)
	log.Log.SetDebugLevel(1)

	s := &Server{
		log: log.Log,
	}

	timer := NewEventTimer(s)
	if timer == nil {
		t.FailNow()
	}

	var receivedTime time.Time
	go func() {
		for {
			select {
			case evt := <-timer.C:
				fmt.Printf("timer received: evt: %d, blkNum: %d \n", evt.evtType, evt.blockNum)
				receivedTime = time.Now()
			case <-timer.stopC:
				break
			}
		}
	}()

	timer.StartTimer(1, time.Millisecond*500)

	time.Sleep(time.Millisecond * 10)
	// double start timer, should overwrite the first one
	startTime := time.Now()
	timer.StartTimer(1, time.Millisecond*100)

	time.Sleep(time.Second)
	timer.Stop()

	if receivedTime.Sub(startTime) > time.Millisecond*150 {
		t.FailNow()
	}
}
