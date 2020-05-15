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
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	data, err := ioutil.ReadFile("vk")
	if err != nil {
		fmt.Printf("ReadFile verifying-key error\n")
		return
	}
	dataLen := len(data)

	vkFileData := "package ontfs\n\nvar vkData = []byte{"

	var tmp string
	for i := 0; i < dataLen; i++ {
		if i%16 == 0 {
			tmp = "\n\t"
		} else {
			tmp = ""
		}
		if i+1 == dataLen {
			tmp += fmt.Sprintf("0x%02x}\n", data[i])
		} else {
			tmp += fmt.Sprintf("0x%02x, ", data[i])
		}

		vkFileData = vkFileData + tmp
	}

	if err = ioutil.WriteFile("vk.go", []byte(vkFileData), 0600); err != nil {
		fmt.Printf("WriteFile vk.go error\n")
		return
	}

	os.Rename("vk.go", "../vk.go")
}
