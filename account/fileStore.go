package account

import (
	ct "DNA/core/contract"
	. "DNA/errors"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
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
		return nil, NewDetailErr(errors.New("[readDB] file handle is nil"), ErrNoCode, "")
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

	return nil, NewDetailErr(errors.New("Can't find the key: "+name), ErrNoCode, "")
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

	cs.fd.ScriptHash = fmt.Sprintf("%x", ct.ProgramHash.ToArray())
	cs.fd.PublicKeyHash = fmt.Sprintf("%x", ct.OwnerPubkeyHash.ToArray())
	cs.fd.RawData = fmt.Sprintf("%x", ct.ToArray())

	jsonblob, err := json.Marshal(cs.fd)
	if err != nil {
		fmt.Println("error:", err)
	}

	cs.writeDB(jsonblob)
	return nil
}
