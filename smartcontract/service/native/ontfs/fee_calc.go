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

const Minute = 60

func calcFileModeRestAmount(timeNow uint64, fileInfo *FileInfo) uint64 {
	fTimeNow := formatUint64TimeToMinute(timeNow)
	fExpired := formatUint64TimeToMinute(fileInfo.TimeExpired)

	if fTimeNow >= fExpired {
		return 0
	}
	restMinute := (fExpired - fTimeNow)/Minute
	return restMinute * fileInfo.CopyNumber * fileInfo.FileBlockCount * fileInfo.CurrFeeRate
}

func calcFileModePerServerProfit(dataClosing uint64, fileInfo *FileInfo) uint64 {
	fStart := formatUint64TimeToMinute(fileInfo.TimeStart)
	fExpired := formatUint64TimeToMinute(fileInfo.TimeExpired)
	dataClosing = formatUint64TimeToMinute(dataClosing)

	if dataClosing <= fStart {
		return 0
	}
	if dataClosing >= fExpired {
		dataClosing = fExpired
	}
	intervalMinute := (dataClosing - fStart)/Minute
	return intervalMinute * fileInfo.FileBlockCount * fileInfo.CurrFeeRate
}

func calcSpaceModePerServerProfit(dataClosing uint64, spaceExpired uint64, fileInfo *FileInfo) uint64 {
	fStart := formatUint64TimeToMinute(fileInfo.TimeStart)
	sExpired := formatUint64TimeToMinute(spaceExpired)
	dataClosing = formatUint64TimeToMinute(dataClosing)

	if dataClosing <= fStart {
		return 0
	}
	if dataClosing < sExpired {
		dataClosing = sExpired
	}
	intervalMinute := (dataClosing - fStart)/Minute
	return intervalMinute * fileInfo.FileBlockCount * fileInfo.CurrFeeRate
}

func calcTotalPayAmountWithFile(fileInfo *FileInfo) uint64 {
	fStart := formatUint64TimeToMinute(fileInfo.TimeStart)
	fExpired := formatUint64TimeToMinute(fileInfo.TimeExpired)
	if fExpired <= fStart {
		return 0
	}
	intervalMinute := (fExpired - fStart)/Minute
	return intervalMinute * fileInfo.CopyNumber * fileInfo.FileBlockCount * fileInfo.CurrFeeRate
}

func calcTotalPayAmountWithSpace(spaceInfo *SpaceInfo) uint64 {
	sStart := formatUint64TimeToMinute(spaceInfo.TimeStart)
	sExpired := formatUint64TimeToMinute(spaceInfo.TimeExpired)
	if sExpired <= sStart {
		return 0
	}
	intervalMinute := (sExpired - sStart)/Minute
	return intervalMinute * spaceInfo.CopyNumber * spaceInfo.Volume * spaceInfo.CurrFeeRate
}
