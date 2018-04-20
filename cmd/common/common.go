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

package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

func EchoJsonDataGracefully(block interface{}) {
	jsons, errs := json.Marshal(block)
	if errs != nil {
		fmt.Printf("Marshal json err:%s", errs.Error())
		return
	}

	var out bytes.Buffer
	err := json.Indent(&out, jsons, "", "\t")
	if err != nil {
		fmt.Printf("Gracefully format json err: %s", err.Error())
		return
	}
	out.WriteTo(os.Stdout)
}
