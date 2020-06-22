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
	"net"
)

//peer capability
const (
	VERIFY_NODE  = 1 //peer involved in consensus
	SERVICE_NODE = 2 //peer only sync with consensus peer
)

const MIN_VERSION_FOR_DHT = "1.9.1-beta"

//link and concurrent const
const (
	PER_SEND_LEN        = 1024 * 256 //byte len per conn write
	MAX_BUF_LEN         = 1024 * 256 //the maximum buffer to receive message
	WRITE_DEADLINE      = 5          //deadline of conn write
	REQ_INTERVAL        = 3          //single request max interval in second
	MAX_REQ_RECORD_SIZE = 1000       //the maximum request record size
	MAX_RESP_CACHE_SIZE = 50         //the maximum response cache
	MAX_TX_CACHE_SIZE   = 100000     //the maximum txHash cache size
)

//msg cmd const
const (
	MSG_CMD_LEN     = 12               //msg type length in byte
	CHECKSUM_LEN    = 4                //checksum length in byte
	MSG_HDR_LEN     = 24               //msg hdr length in byte
	MAX_BLK_HDR_CNT = 500              //hdr count once when sync header
	MAX_MSG_LEN     = 30 * 1024 * 1024 //the maximum message length
	MAX_PAYLOAD_LEN = MAX_MSG_LEN - MSG_HDR_LEN
)

//msg type const
const (
	MAX_ADDR_NODE_CNT = 64 //the maximum peer address from msg
	MAX_INV_BLK_CNT   = 64 //the maximum blk hash cnt of inv msg
)

//info update const
const (
	PROTOCOL_VERSION      = 0     //protocol version
	UPDATE_RATE_PER_BLOCK = 2     //info update rate in one generate block period
	KEEPALIVE_TIMEOUT     = 15    //contact timeout in sec
	DIAL_TIMEOUT          = 6     //connect timeout in sec
	CONN_MONITOR          = 6     //time to retry connect in sec
	CONN_MAX_BACK         = 4000  //max backoff time in micro sec
	MAX_RETRY_COUNT       = 3     //max reconnect time of remote peer
	CHAN_CAPABILITY       = 10000 //channel capability of recv link
	SYNC_BLK_WAIT         = 2     //timespan for blk sync check
)

const (
	RecentPeerElapseLimit = 60
)

//cap flag
const HTTP_INFO_FLAG = 0 //peer`s http info bit in cap field

//recent contact const
const (
	RECENT_TIMEOUT   = 60
	RECENT_FILE_NAME = "peers.recent"
)

//PeerAddr represent peer`s net information
type PeerAddr struct {
	Time     int64    //latest timestamp
	Services uint64   //service type
	IpAddr   [16]byte //ip address
	Port     uint16   //sync port
	//todo remove this legecy field
	ConsensusPort uint16 //consensus port
	ID            PeerId //Unique ID
}

//const channel msg id and type
const (
	VERSION_TYPE       = "version"     //peer`s information
	VERACK_TYPE        = "verack"      //ack msg after version recv
	GetADDR_TYPE       = "getaddr"     //req nbr address from peer
	ADDR_TYPE          = "addr"        //nbr address
	PING_TYPE          = "ping"        //ping  sync height
	PONG_TYPE          = "pong"        //pong  recv nbr height
	GET_HEADERS_TYPE   = "getheaders"  //req blk hdr
	HEADERS_TYPE       = "headers"     //blk hdr
	INV_TYPE           = "inv"         //inv payload
	GET_DATA_TYPE      = "getdata"     //req data from peer
	BLOCK_TYPE         = "block"       //blk payload
	TX_TYPE            = "tx"          //transaction
	CONSENSUS_TYPE     = "consensus"   //consensus payload
	GET_BLOCKS_TYPE    = "getblocks"   //req blks from peer
	NOT_FOUND_TYPE     = "notfound"    //peer can`t find blk according to the hash
	FINDNODE_TYPE      = "findnode"    // find node using dht
	FINDNODE_RESP_TYPE = "findnodeack" // find node using dht
	UPDATE_KADID_TYPE  = "updatekadid" //update node kadid

	GET_SUBNET_MEMBERS_TYPE = "getmembers" // request subnet members
	SUBNET_MEMBERS_TYPE     = "members"    // response subnet members
)

//ParseIPAddr return ip address
func ParseIPAddr(s string) (string, error) {
	host, _, err := net.SplitHostPort(s)
	return host, err
}
