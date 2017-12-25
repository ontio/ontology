package states

type CoinState byte

const (
	Unconfirmed CoinState = iota
	Confirmed
	Spent
)