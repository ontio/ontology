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
