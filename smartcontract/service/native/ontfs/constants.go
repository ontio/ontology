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

package ontfs

const (
	DefaultPerBlockSize = 256 //kb.
)

const (
	DefaultMinTimeForFileStorage = 60 * 60 * 24 //1day
	DefaultContractInvokeGasFee  = 10000000     //0.01ong
	DefaultChallengeReward       = 100000000    //0.1ong
	DefaultFilePerServerPdpTimes = 2
	DefaultPassportExpire        = 9           //block count. passport expire for GetFileHashList
	DefaultNodeMinVolume         = 1024 * 1024 //kb. min total volume with single fsNode
	DefaultChallengeInterval     = 1 * 60 * 60 //1hour
	DefaultNodePerKbPledge       = 1024 * 100  //fsNode's pledge for participant
	DefaultFilePerBlockFeeRate   = 60          //file mode cost of per block save from fsNode for one minute
	DefaultSpacePerBlockFeeRate  = 60          //space mode cost of per block save from fsNode for one hour
	DefaultGasPerBlockForRead    = 256         //cost of per block read from fsNode
)

//challenge state
const (
	Judged = iota
	NoReplyAndValid
	NoReplyAndExpire
	RepliedAndSuccess
	RepliedButVerifyError
)
