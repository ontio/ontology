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

package account

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	ct "github.com/ontio/ontology/core/contract"
	ontErrors "github.com/ontio/ontology/errors"
)

type FileData struct {
	PublicKeyHash       string
	PrivateKeyEncrypted string
	Address             string
	ScriptHash          string
	RawData             string
	PasswordHash        string
	IV                  string
	MasterKey           string
}

type FileStore struct {
	fd   FileData
	file *os.File
	path string
}

func (cs *FileStore) readDB() ([]byte, error) {
	var err error
	cs.file, err = os.OpenFile(cs.path, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer cs.closeDB()

	if cs.file != nil {
		data, err := ioutil.ReadAll(cs.file)
		if err != nil {
			return nil, err
		}

		return data, nil

	} else {
		return nil, ontErrors.NewDetailErr(errors.New("[readDB] file handle is nil"), ontErrors.ErrNoCode, "")
	}
}

func (cs *FileStore) writeDB(data []byte) error {
	var err error
	cs.file, err = os.OpenFile(cs.path, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer cs.closeDB()

	if cs.file != nil {
		cs.file.Write(data)
	}

	return nil
}

func (cs *FileStore) closeDB() {
	if cs.file != nil {
		cs.file.Close()
		cs.file = nil
	}
}

func (cs *FileStore) BuildDatabase(path string) {
	err := os.Remove(path)
	if err != nil {
		//FIXME ignore this error
	}

	jsonBlob := []byte("{\"PublicKeyHash\":\"\", \"PrivateKeyEncrypted\":\"\", \"Address\":\"\", \"ScriptHash\":\"\", \"RawData\":\"\", \"PasswordHash\":\"\", \"IV\":\"\", \"MasterKey\":\"\"}")

	cs.writeDB(jsonBlob)
}

func (cs *FileStore) SaveStoredData(name string, value []byte) error {
	jsondata, err := cs.readDB()
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsondata, &cs.fd)
	if err != nil {
		fmt.Println("error:", err)
	}

	if name == "IV" {
		cs.fd.IV = fmt.Sprintf("%x", value)
	} else if name == "MasterKey" {
		cs.fd.MasterKey = fmt.Sprintf("%x", value)
	} else if name == "PasswordHash" {
		cs.fd.PasswordHash = fmt.Sprintf("%x", value)
	}

	jsonblob, err := json.Marshal(cs.fd)
	if err != nil {
		fmt.Println("error:", err)
	}

	cs.writeDB(jsonblob)

	return nil
}

func (cs *FileStore) LoadStoredData(name string) ([]byte, error) {
	jsondata, err := cs.readDB()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsondata, &cs.fd)
	if err != nil {
		fmt.Println("error:", err)
	}

	if name == "IV" {
		return hex.DecodeString(cs.fd.IV)
	} else if name == "MasterKey" {
		return hex.DecodeString(cs.fd.MasterKey)
	} else if name == "PasswordHash" {
		return hex.DecodeString(cs.fd.PasswordHash)
	}

	return nil, ontErrors.NewDetailErr(errors.New("Can't find the key: "+name), ontErrors.ErrNoCode, "")
}

func (cs *FileStore) SaveAccountData(pubkeyhash []byte, prikeyenc []byte) error {
	jsondata, err := cs.readDB()
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsondata, &cs.fd)
	if err != nil {
		fmt.Println("error:", err)
	}

	cs.fd.PublicKeyHash = fmt.Sprintf("%x", pubkeyhash)
	cs.fd.PrivateKeyEncrypted = fmt.Sprintf("%x", prikeyenc)

	jsonblob, err := json.Marshal(cs.fd)
	if err != nil {
		fmt.Println("error:", err)
	}

	cs.writeDB(jsonblob)
	return nil
}

func (cs *FileStore) LoadAccountData(index int) ([]byte, []byte, error) {
	jsondata, err := cs.readDB()
	if err != nil {
		return nil, nil, err
	}

	err = json.Unmarshal(jsondata, &cs.fd)
	if err != nil {
		fmt.Println("error:", err)
	}

	publickeyHash, err := hex.DecodeString(cs.fd.PublicKeyHash)
	privatekeyEncrypted, err := hex.DecodeString(cs.fd.PrivateKeyEncrypted)

	return publickeyHash, privatekeyEncrypted, err
}

func (cs *FileStore) LoadContractData(index int) ([]byte, []byte, []byte, error) {
	jsondata, err := cs.readDB()
	if err != nil {
		return nil, nil, nil, err
	}

	err = json.Unmarshal(jsondata, &cs.fd)
	if err != nil {
		fmt.Println("error:", err)
	}

	scriptHash, err := hex.DecodeString(cs.fd.ScriptHash)
	publickeyHash, err := hex.DecodeString(cs.fd.PublicKeyHash)
	rawData, err := hex.DecodeString(cs.fd.RawData)

	return scriptHash, publickeyHash, rawData, err
}

func (cs *FileStore) SaveContractData(ct *ct.Contract) error {
	jsondata, err := cs.readDB()
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsondata, &cs.fd)
	if err != nil {
		fmt.Println("error:", err)
	}

	cs.fd.ScriptHash = fmt.Sprintf("%x", ct.ProgramHash[:])
	cs.fd.PublicKeyHash = fmt.Sprintf("%x", ct.OwnerPubkeyHash[:])
	cs.fd.RawData = fmt.Sprintf("%x", ct.ToArray())

	jsonblob, err := json.Marshal(cs.fd)
	if err != nil {
		fmt.Println("error:", err)
	}

	cs.writeDB(jsonblob)
	return nil
}
