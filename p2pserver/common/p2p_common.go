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

package common

import (
	"github.com/ontio/ontology/core/types"
)

//peer capability
const (
	VERIFY_NODE  = 1 //peer involved in consensus
	SERVICE_NODE = 2 //peer only sync with consensus peer
)

//msg cmd const
const (
	MSG_CMD_LEN      = 12         //msg type length in byte
	CMD_OFFSET       = 4          //cmd type offet in msg hdr
	CHECKSUM_LEN     = 4          //checksum length in byte
	HASH_LEN         = 32         // hash length in byte
	MSG_HDR_LEN      = 24         //msg hdr length in byte
	MAX_BLK_HDR_CNT  = 500        //hdr count once when sync header
	MAX_INV_HDR_CNT  = 500        //inventory count once when req inv
	NETMAGIC         = 0x74746e41 //network magic number
	MAX_REQ_BLK_ONCE = 16         //req blk count once from one peer when sync blk

)

//info update const
const (
	PROTOCOL_VERSION      = 0         //protocol version
	MAX_BUF_LEN           = 1024 * 16 //the maximum buffer to receive message
	UPDATE_RATE_PER_BLOCK = 2         //info update rate in one generate block period
	KEEPALIVE_TIMEOUT     = 15        //contact timeout in sec
	DIAL_TIMEOUT          = 6         //connect timeout in sec
	CONN_MONITOR          = 6         //time to retry connect in sec
	CONN_MAX_BACK         = 4000      //max backoff time in micro sec
	MAX_RETRY_COUNT       = 3         //max reconnect time of remote peer
	CHAN_CAPABILITY       = 10000     //channel capability of recv link
	SYNC_BLK_WAIT         = 2         //timespan for blk sync check
)

// The peer state
const (
	INIT        = 0 //initial
	HAND        = 1 //send verion to peer
	HAND_SHAKE  = 2 //haven`t send verion to peer and receive peer`s version
	HAND_SHAKED = 3 //send verion to peer and receive peer`s version
	ESTABLISH   = 4 //receive peer`s verack
	INACTIVITY  = 5 //link broken
)

//cap flag
const (
	HTTP_INFO_FLAG = 0 //peer`s http info bit in cap field
)

//actor const
const (
	ACTOR_TIMEOUT = 5 //actor request timeout in secs
)

//PeerAddr represent peer`s net information
type PeerAddr struct {
	Time          int64    //latest timestamp
	Services      uint64   //service type
	IpAddr        [16]byte //ip address
	Port          uint16   //sync port
	ConsensusPort uint16   //consensus port
	ID            uint64   // Unique ID
}

//MsgPayload in link channel
type MsgPayload struct {
	Id      uint64 //peer ID
	Addr    string //link address
	Payload []byte //msg payload
}

//const channel msg id and type
const (
	VERSION_TYPE     = "version"    //peer`s information
	VERACK_TYPE      = "verack"     //ack msg after version recv
	GetADDR_TYPE     = "getaddr"    //req nbr address from peer
	ADDR_TYPE        = "addr"       //nbr address
	PING_TYPE        = "ping"       //ping  sync height
	PONG_TYPE        = "pong"       //pong  recv nbr height
	GET_HEADERS_TYPE = "getheaders" //req blk hdr
	HEADERS_TYPE     = "headers"    //blk hdr
	INV_TYPE         = "inv"        //inv payload
	GET_DATA_TYPE    = "getdata"    //req data from peer
	BLOCK_TYPE       = "block"      //blk payload
	TX_TYPE          = "tx"         //transaction
	CONSENSUS_TYPE   = "consensus"  //consensus payload
	GET_BLOCKS_TYPE  = "getblocks"  //req blks from peer
	NOT_FOUND_TYPE   = "notfound"   //peer can`t find blk according to the hash
	DISCONNECT_TYPE  = "disconnect" //peer disconnect info raise by link
)

type AppendPeerID struct {
	ID uint64 // The peer id
}

type RemovePeerID struct {
	ID uint64 // The peer id
}

type AppendHeaders struct {
	Headers []*types.Header // Headers to be added to the ledger
}

type AppendBlock struct {
	Block *types.Block // Block to be added to the ledger
}
