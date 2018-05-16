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

package oracle

import "math/big"

type Status int

type RegisterOracleNodeParam struct {
	Address  string `json:"address"`
	Guaranty uint64 `json:"guaranty"`
}

type ApproveOracleNodeParam struct {
	Address string `json:"address"`
}

type OracleNode struct {
	Address  string `json:"address"`
	Guaranty uint64 `json:"guaranty"`
	Status   Status `json:"status"`
}

type QuitOracleNodeParam struct {
	Address string `json:"address"`
}

type CreateOracleRequestParam struct {
	Request   string   `json:"request"`
	OracleNum *big.Int `json:"oracleNum"`
	Address   string   `json:"address"`
}

type UndoRequests struct {
	Requests map[string]struct{} `json:"requests"`
}

type SetOracleOutcomeParam struct {
	TxHash  string      `json:"txHash"`
	Address string      `json:"owner"`
	Outcome interface{} `json:"outcome"`
}

type OutcomeRecord struct {
	OutcomeRecord map[string]interface{} `json:"outcomeRecord"`
}

type SetOracleCronOutcomeParam struct {
	TxHash  string      `json:"txHash"`
	Address string      `json:"owner"`
	Outcome interface{} `json:"outcome"`
}

type CronOutcomeRecord struct {
	CronOutcomeRecord map[string]interface{} `json:"cronOutcomeRecord"`
}

type ChangeCronViewParam struct {
	TxHash  string `json:"txHash"`
	Address string `json:"owner"`
}
