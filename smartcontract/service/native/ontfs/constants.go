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
	DefaultPassportExpire = 9 //block count. passport expire for GetFileHashList

	DefaultNodeMinVolume   = 1024 * 1024 //kb. min total volume with fsNode
	DefaultNodePerKbPledge = 1           //fsNode's pledge for participant

	DefaultMinFileStoreTime         = 4 * 60 * 60
	DefaultMinDownLoadFee           = 1 //min download fee for single task*
	DefaultGasPerKbForRead          = 1 //cost for ontfs-sdk read from fsNode*
	DefaultGasPerKbForSaveWithFile  = 1 //cost for ontfs-sdk save from fsNode*
	DefaultGasPerKbForSaveWithSpace = 1 //cost for ontfs-sdk save from fsNode*

	DefaultChallengeInterval = 4 * 60 * 60
	DefaultPdpHeightIV       = 8   //pdp challenge height IV
	DefaultPerBlockSize      = 256 //kb.
	DefaultPdpBlockNum       = 32
)

//challenge state
const (
	Judged = iota
	NoReplyAndValid
	NoReplyAndExpire
	RepliedAndSuccess
	RepliedButVerifyError
	FileProveSuccess
)
