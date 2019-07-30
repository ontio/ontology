package wasmvm

func min(x, y uint64) uint64 {
	if x < y {
		return x
	}
	return y
}
