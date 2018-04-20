package states

import "math/big"

type CreateOracleRequestParam struct {
	Request   string   `json:"request"`
	OracleNum *big.Int `json:"oracleNum"`
	Address   string   `json:"address"`
}

type UndoRequests struct {
	Requests map[string]interface{} `json:"requests"`
}

type SetOracleOutcomeParam struct {
	TxHash  string      `json:"txHash"`
	Owner   string      `json:"owner"`
	Outcome interface{} `json:"outcome"`
}

type OutcomeRecord struct {
	OutcomeRecord map[string]interface{} `json:"outcomeRecord"`
}

type SetOracleCronOutcomeParam struct {
	TxHash  string      `json:"txHash"`
	Owner   string      `json:"owner"`
	Outcome interface{} `json:"outcome"`
}

type CronOutcomeRecord struct {
	CronOutcomeRecord map[string]interface{} `json:"cronOutcomeRecord"`
}

type ChangeCronViewParam struct {
	TxHash string `json:"txHash"`
	Owner  string `json:"owner"`
}
