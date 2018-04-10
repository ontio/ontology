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

//peer capability
const (
	VERIFY_NODE  = 1
	SERVICE_NODE = 2
)

//config const
const (
	VERIFY_NODE_NAME  = "verify"
	SERVICE_NODE_NAME = "service"
)

//msg cmd const
const (
	MSG_CMD_LEN           = 12
	CMD_OFFSET            = 4
	CHECKSUM_LEN          = 4
	HASH_LEN              = 32 // hash length in byte
	MSG_HDR_LEN           = 24
	MAX_BLK_HDR_CNT       = 500
	MAX_INV_HDR_CNT       = 500
	NETMAGIC              = 0x74746e41
	DIV_HASH_LEN          = 5
	MAX_REQ_BLK_ONCE      = 16
	UPDATE_RATE_PER_BLOCK = 2
)

//info update const
const (
	PROTOCOL_VERSION   = 0
	HELLO_TIMEOUT      = 3 // Seconds
	MAX_HELLO_RETYR    = 3
	MAX_BUF_LEN        = 1024 * 16 // Fixme The maximum buffer to receive message
	MAX_CHAN_BUF       = 512
	PERIOD_UPDATE_TIME = 3 // Time to update and sync information with other nodes
	HEARTBEAT          = 2
	KEEPALIVE_TIMEOUT  = 3
	DIAL_TIMEOUT       = 6
	CONN_MONITOR       = 6
	CONN_MAX_BACK      = 4000
	MAX_RETRY_COUNT    = 3
)

// The peer state
const (
	INIT       = 0
	HAND       = 1
	HANDSHAKE  = 2
	HANDSHAKED = 3
	ESTABLISH  = 4
	INACTIVITY = 5
)

//cap flag
const (
	HTTP_INFO_FLAG = 0
)

//PeerAddr represent peer`s net information
type PeerAddr struct {
	Time          int64
	Services      uint64
	IpAddr        [16]byte
	Port          uint16
	ConsensusPort uint16
	ID            uint64 // Unique ID
}

//MsgPayload in link channel
type MsgPayload struct {
	Id      uint64 //peer ID
	Addr    string //link address
	Payload []byte
}

//const channel msg id and type
const (
	VERSION_TYPE      = "version"
	VERACK_TYPE       = "verack"
	GetADDR_TYPE      = "getaddr"
	ADDR_TYPE         = "addr"
	PING_TYPE         = "ping"
	PONG_TYPE         = "pong"
	GET_HEADERS_TYPE  = "getheaders"
	HEADERS_TYPE      = "headers"
	INV_TYPE          = "inv"
	GET_DATA_TYPE     = "getdata"
	BLOCK_TYPE        = "block"
	TX_TYPE           = "tx"
	CONSENSUS_TYPE    = "consensus"
	FILTER_ADD_TYPE   = "filteradd"
	FILTER_CLEAR_TYPE = "filterclear"
	FILTER_LOAD_TYPE  = "filterload"
	GET_BLOCKS_TYPE   = "getblocks"
	NOT_FOUND_TYPE    = "notfound"
	DISCONNECT_TYPE   = "disconnect"
)
