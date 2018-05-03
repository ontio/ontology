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

package cmd

import (
	"github.com/ontio/ontology/cmd/utils"
	"github.com/urfave/cli"
)

var (
	NodeFlags = []cli.Flag{
		utils.LogLevelFlag,
		//utils.ConsensusFlag,
		utils.WalletFileFlag,
		utils.AccountPassFlag,
	}

	ContractFlags = []cli.Flag{
		utils.ContractVmTypeFlag,
		utils.ContractStorageFlag,
		utils.ContractCodeFileFlag,
		utils.ContractNameFlag,
		utils.ContractVersionFlag,
		utils.ContractAuthorFlag,
		utils.ContractEmailFlag,
		utils.ContractDescFlag,
		utils.ContractParamsFlag,
		utils.ContractAddrFlag,
	}

	InfoFlags = []cli.Flag{
		utils.BlockHeightInfoFlag,
		utils.BlockHashInfoFlag,
	}

	listFlags = []cli.Flag{
		utils.AccountVerboseFlag,
		utils.WalletFileFlag,
	}

	setFlags = []cli.Flag{
		utils.AccountSigSchemeFlag,
		utils.AccountSetDefaultFlag,
		utils.WalletFileFlag,
	}
	addFlags = []cli.Flag{
		utils.AccountTypeFlag,
		utils.AccountKeylenFlag,
		utils.AccountSigSchemeFlag,
		utils.AccountPassFlag,
		utils.AccountDefaultFlag,
		utils.WalletFileFlag,
	}
	fileFlags = []cli.Flag{
		utils.WalletFileFlag,
	}
)
