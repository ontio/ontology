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
	sdk "github.com/ontio/ontology-go-sdk"
)

var ontSdk *sdk.OntologySdk

func rpcAddress() string {
	//index := rand.Intn(len(config.Parameters.SeedList))
	//return (strings.Split(config.Parameters.SeedList[index], ":"))[0] + strconv.Itoa(config.Parameters.HttpJsonPort)
	return "http://139.219.108.204:20336"
}

func init() {
	ontSdk = sdk.NewOntologySdk()
	ontSdk.Rpc.SetAddress(rpcAddress())
}
