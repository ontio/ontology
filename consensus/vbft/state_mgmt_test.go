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

package vbft

import (
	"testing"
	"time"

	"github.com/ontio/ontology/common/log"
)

func Test_isReady(t *testing.T) {
	type args struct {
		state ServerState
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test",
			args: args{state: Synced},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isReady(tt.args.state); got != tt.want {
				t.Errorf("isReady() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isActive(t *testing.T) {
	type args struct {
		state ServerState
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test",
			args: args{state: SyncReady},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isActive(tt.args.state); got != tt.want {
				t.Errorf("isActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStateMgr_getState(t *testing.T) {
	sev := constructServer()
	type fields struct {
		server              *Server
		syncReadyTimeout    time.Duration
		currentState        ServerState
		StateEventC         chan *StateEvent
		peers               map[uint32]*PeerState
		liveTicker          *time.Timer
		lastTickChainHeight uint32
	}
	tests := []struct {
		name   string
		fields fields
		want   ServerState
	}{
		{
			name:   "test",
			fields: fields{server: sev, syncReadyTimeout: 5, currentState: Syncing},
			want:   Syncing,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self := &StateMgr{
				server:              tt.fields.server,
				syncReadyTimeout:    tt.fields.syncReadyTimeout,
				currentState:        tt.fields.currentState,
				StateEventC:         tt.fields.StateEventC,
				peers:               tt.fields.peers,
				liveTicker:          tt.fields.liveTicker,
				lastTickChainHeight: tt.fields.lastTickChainHeight,
			}
			if got := self.getState(); got != tt.want {
				t.Errorf("StateMgr.getState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStateMgr_onPeerUpdate(t *testing.T) {
	log.Init(log.PATH, log.Stdout)
	sev := constructServer()
	peerstate := &PeerState{
		peerIdx:           1,
		chainConfigView:   0,
		committedBlockNum: 1,
		connected:         true,
	}
	peers := make(map[uint32]*PeerState)
	peers[1] = peerstate
	type fields struct {
		server              *Server
		syncReadyTimeout    time.Duration
		currentState        ServerState
		StateEventC         chan *StateEvent
		peers               map[uint32]*PeerState
		liveTicker          *time.Timer
		lastTickChainHeight uint32
	}
	type args struct {
		peerState *PeerState
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{server: sev, syncReadyTimeout: 2, currentState: Syncing, StateEventC: make(chan *StateEvent, 16),
				peers: peers},
			args:    args{peerState: peerstate},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self := &StateMgr{
				server:              tt.fields.server,
				syncReadyTimeout:    tt.fields.syncReadyTimeout,
				currentState:        tt.fields.currentState,
				StateEventC:         tt.fields.StateEventC,
				peers:               tt.fields.peers,
				liveTicker:          tt.fields.liveTicker,
				lastTickChainHeight: tt.fields.lastTickChainHeight,
			}
			if err := self.onPeerUpdate(tt.args.peerState); (err != nil) != tt.wantErr {
				t.Errorf("StateMgr.onPeerUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func constructPeerState() *StateMgr {
	sev := constructServer()
	statemgr := newStateMgr(sev)
	peerstate := &PeerState{
		peerIdx:           1,
		chainConfigView:   0,
		committedBlockNum: 2,
		connected:         true,
	}
	peers := make(map[uint32]*PeerState)
	peers[peerstate.peerIdx] = peerstate
	statemgr.peers = peers
	return statemgr
}
func TestStateMgr_onPeerDisconnected(t *testing.T) {
	statemgr := constructPeerState()
	err := statemgr.onPeerDisconnected(1)
	if err != nil {
		t.Errorf("TestonPeerDisconnected failed:%v", err)
		return
	}
	t.Logf("TestonPeerDisconnected succ")
}

func TestStateMgr_onLiveTick(t *testing.T) {
	statemgr := constructPeerState()
	statemgr.lastTickChainHeight = 1
	peerstate := &PeerState{
		peerIdx:           1,
		chainConfigView:   0,
		committedBlockNum: 1,
		connected:         true,
	}
	stateevent := &StateEvent{
		Type:      SyncDone,
		peerState: peerstate,
		blockNum:  1,
	}
	err := statemgr.onLiveTick(stateevent)
	if err != nil {
		t.Errorf("TestonLiveTick failed: %v", err)
		return
	}
	t.Logf("TestonLiveTick succ")
}

func TestStateMgr_getSyncedChainConfigView(t *testing.T) {
	statemgr := constructPeerState()
	statemgr.lastTickChainHeight = 1
	viewnum := statemgr.getSyncedChainConfigView()
	t.Logf("TestgetSyncedChainConfigView  view:%d", viewnum)
}

func TestStateMgr_isSyncedReady(t *testing.T) {
	statemgr := constructPeerState()
	statemgr.lastTickChainHeight = 1
	flag := statemgr.isSyncedReady()
	t.Logf("TestisSyncedReady %v", flag)
}

func TestStateMgr_checkStartSyncing(t *testing.T) {
	log.Init(log.PATH, log.Stdout)
	statemgr := constructPeerState()
	statemgr.server.syncer = newSyncer(statemgr.server)
	statemgr.checkStartSyncing(1, true)
	t.Log("TestcheckStartSyncing")
}

func TestStateMgr_getConsensusedCommittedBlockNum(t *testing.T) {
	log.Init(log.PATH, log.Stdout)
	statemgr := constructPeerState()
	maxcomit, flag := statemgr.getConsensusedCommittedBlockNum()
	t.Logf("TestgetConsensusedCommittedBlockNum maxcommitted:%v, consensused:%v", maxcomit, flag)
}
