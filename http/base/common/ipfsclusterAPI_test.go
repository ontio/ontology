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
	"fmt"
	"github.com/Ontology/common/log"
	"os"
	"os/exec"
	"testing"
)

func TestAddFileIPFS(t *testing.T) {
	var path string = "./Log/"
	log.CreatePrintLog(path)
	cmd := exec.Command("/bin/sh", "-c", "dd if=/dev/zero of=test bs=1024 count=1000")
	cmd.Run()
	ref, err := AddFileIPFS("test", true)
	if err != nil {
		t.Fatalf("AddFileIPFS error:%s", err.Error())
	}
	os.Remove("test")
	fmt.Printf("ipfs path=%s\n", ref)
}
func TestGetFileIPFS(t *testing.T) {
	var path string = "./Log/"
	log.CreatePrintLog(path)
	ref := "QmVHzLjYvp4bposJDD2PNeJ9PAFixyQu3oFj6gqipgsukX"
	err := GetFileIPFS(ref, "testOut")
	if err != nil {
		t.Fatalf("GetFileIPFS error:%s", err.Error())
	}
	//os.Remove("testOut")
}
