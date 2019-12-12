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

package main

import (
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/wasmtest/common"
)

const (
	testdir = "testwasmdata"
)

func main() {
	acct, database := common.InitOntologyLedger()

	log.Info("loading contract")
	contract, objIsDir, err := common.GetContact(testdir)
	if err != nil {
		panic(err)
	}
	if !objIsDir {
		panic("testwasmdata is not a dir")
	}

	log.Infof("deploying %d contracts", len(contract))
	err = common.DeployContract(acct, database, contract)
	if err != nil {
		panic(err)
	}

	testContext := common.MakeTestContext(acct, contract)

	common.TestWithbatchMode(acct, database, contract, testContext)
}
