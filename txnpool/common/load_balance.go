package common

/*
 * Implement sort.Interface
 */
type LoadBalance struct {
	Size     int
	WorkerID uint8
}

type LoadBalances []LoadBalance

func (this LoadBalances) Len() int {
	return len(this)
}

func (this LoadBalances) Swap(i, j int) {
	this[i].Size, this[j].Size = this[j].Size, this[i].Size
	this[i].WorkerID, this[j].WorkerID = this[j].WorkerID, this[i].WorkerID
}

func (this LoadBalances) Less(i, j int) bool {
	return this[i].Size < this[j].Size
}
