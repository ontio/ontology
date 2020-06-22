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

const Hour = 3600

func calcFileModeRestAmount(timeNow uint64, fileInfo *FileInfo) uint64 {
	fTimeNow := formatUint64TimeToHour(timeNow)
	fExpired := formatUint64TimeToHour(fileInfo.TimeExpired)

	if fTimeNow >= fExpired {
		return 0
	}
	restHour := (fExpired - fTimeNow) / Hour
	return restHour * fileInfo.CopyNumber * fileInfo.FileBlockCount * fileInfo.CurrFeeRate
}

func calcFileModePerServerProfit(dataClosing uint64, fileInfo *FileInfo) uint64 {
	fStart := formatUint64TimeToHour(fileInfo.TimeStart)
	fExpired := formatUint64TimeToHour(fileInfo.TimeExpired)
	dataClosing = formatUint64TimeToHour(dataClosing)

	if dataClosing <= fStart {
		return 0
	}
	if dataClosing >= fExpired {
		dataClosing = fExpired
	}
	intervalHour := (dataClosing - fStart) / Hour
	return intervalHour * fileInfo.FileBlockCount * fileInfo.CurrFeeRate
}

func calcSpaceModePerServerProfit(dataClosing uint64, spaceExpired uint64, fileInfo *FileInfo) uint64 {
	fStart := formatUint64TimeToHour(fileInfo.TimeStart)
	sExpired := formatUint64TimeToHour(spaceExpired)
	dataClosing = formatUint64TimeToHour(dataClosing)

	if dataClosing <= fStart {
		return 0
	}
	if dataClosing < sExpired {
		dataClosing = sExpired
	}
	intervalHour := (dataClosing - fStart) / Hour
	return intervalHour * fileInfo.FileBlockCount * fileInfo.CurrFeeRate
}

func calcTotalPayAmountWithFile(fileInfo *FileInfo) uint64 {
	fStart := formatUint64TimeToHour(fileInfo.TimeStart)
	fExpired := formatUint64TimeToHour(fileInfo.TimeExpired)
	if fExpired <= fStart {
		return 0
	}
	intervalHour := (fExpired - fStart) / Hour
	return intervalHour * fileInfo.CopyNumber * fileInfo.FileBlockCount * fileInfo.CurrFeeRate
}

func calcTotalPayAmountWithSpaceFile(fileInfo *FileInfo, spaceTimeExpired uint64) uint64 {
	fStart := formatUint64TimeToHour(fileInfo.TimeStart)
	fExpired := formatUint64TimeToHour(spaceTimeExpired)
	if fExpired <= fStart {
		return 0
	}
	intervalHour := (fExpired - fStart) / Hour
	return intervalHour * fileInfo.CopyNumber * fileInfo.FileBlockCount * fileInfo.CurrFeeRate
}

func calcTotalPayAmountWithSpace(spaceInfo *SpaceInfo) uint64 {
	sStart := formatUint64TimeToHour(spaceInfo.TimeStart)
	sExpired := formatUint64TimeToHour(spaceInfo.TimeExpired)
	if sExpired <= sStart {
		return 0
	}
	intervalHour := (sExpired - sStart) / Hour
	return intervalHour * spaceInfo.CopyNumber * (spaceInfo.Volume / 256) * spaceInfo.CurrFeeRate
}
