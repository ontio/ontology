package client

import (
	ct "GoOnchain/core/contract"
	//. "GoOnchain/common"
	"fmt"
	"os"
	. "GoOnchain/errors"
	"errors"
	"encoding/json"
	"io/ioutil"
	"encoding/hex"
)

type FileData struct {
	PublicKeyHash  string
	PrivateKeyEncrypted string
	Address string
	ScriptHash string
	RawData string
	PasswordHash string
	IV string
	MasterKey string
}

type FileStore struct {
	fd FileData
	file *os.File
	path string
}

func (cs * FileStore) readDB() ([]byte,error) {
	var err error
	cs.file, err = os.OpenFile(cs.path, os.O_RDONLY, 0666)
	if err != nil {
		return nil,err
	}
	defer cs.closeDB()

	if cs.file != nil {
		data,err := ioutil.ReadAll(cs.file)
		if err != nil {
			return nil,err
		}

		return data,nil

	} else {
		return nil,NewDetailErr(errors.New("[readDB] file handle is nil"), ErrNoCode, "")
	}
}

func (cs * FileStore) writeDB( data []byte ) error {

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

func (cs * FileStore) closeDB() {
	if  cs.file != nil {
		cs.file.Close()
		cs.file = nil
	}
}

func (cs * FileStore) BuildDatabase(path string) {

	err := os.Remove(path)
	if err != nil {
		//fmt.Println( err )
		//return
	}

	jsonBlob := []byte("{\"PublicKeyHash\":\"\", \"PrivateKeyEncrypted\":\"\", \"Address\":\"\", \"ScriptHash\":\"\", \"RawData\":\"\", \"PasswordHash\":\"\", \"IV\":\"\", \"MasterKey\":\"\"}")
	//fmt.Printf("FileData json %s\n",string(jsonBlob))

	cs.writeDB( jsonBlob )
}

func (cs * FileStore) SaveStoredData(name string,value []byte) error {

	jsondata,err := cs.readDB()
	if err != nil {
		return err
	}
	//fmt.Printf("%s\n",jsondata)

	err = json.Unmarshal(jsondata, &cs.fd)
	if err != nil {
		fmt.Println("error:", err)
	}

	if name == "IV" {
		cs.fd.IV = fmt.Sprintf("%x",value)
	} else if name == "MasterKey" {
		cs.fd.MasterKey = fmt.Sprintf("%x",value)
	} else if name == "PasswordHash" {
		cs.fd.PasswordHash = fmt.Sprintf("%x",value)
	}

	jsonblob, err := json.Marshal(cs.fd)
	if err != nil {
		fmt.Println("error:", err)
	}

	cs.writeDB( jsonblob )

	/*
	err := cs.OpenDB(cs.path)
	defer cs.CloseDB()
	if err != nil {
		// TODO: error
		return err
	}*/
/*
	//select key name
	sql := "SELECT * FROM Key WHERE name = '" + name + "'"
	rows, err := cs.db.Query( sql )
	if err != nil {
		return err
	}

	// check if name exist
	if rows.Next() {
		// update
		stmt, err := cs.db.Prepare("UPDATE key SET Value=? where Name=?")
		if err != nil {
			return err
		}

		res, err := stmt.Exec(value, name)
		if err != nil {
			return err
		}

		affect, err := res.RowsAffected()
		if err != nil {
			return err
		}

		fmt.Printf("UPDATE Key: [%s] %x Index:%d\n",  name, value, affect )

	} else{
		// insert
		stmt, err := cs.db.Prepare("INSERT INTO Key(Name,Value) values(?,?)")
		if err != nil {
			return err
		}

		res, err := stmt.Exec(name, value)
		if err != nil {
			return err
		}

		_, err = res.LastInsertId()
		if err != nil {
			return err
		}

		//fmt.Printf("INSERT Key: [%s] %x Index:%d\n",  name, value, id )
	}
*/
	return nil
}

func (cs * FileStore) LoadStoredData(name string) ([]byte,error) {

	jsondata,err := cs.readDB()
	if err != nil {
		return nil,err
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

/*
	err := cs.OpenDB(cs.path)
	defer cs.CloseDB()
	if err != nil {
		// TODO: error
		return nil,err
	}
*/
/*
	sql := "SELECT * FROM Key WHERE name = '" + name + "'"
	rows, err := cs.db.Query( sql )
	if err != nil {
		return nil,err
	}

	// check if name exist
	for rows.Next() {
		var n string
		var v []byte

		rows.Scan(&n, &v)

		if strings.EqualFold( n, name ) {
			return v, nil
		}
	}
*/
	return nil, NewDetailErr(errors.New("Can't find the key: " + name), ErrNoCode, "")
}

func (cs * FileStore) SaveAccountData(pubkeyhash []byte,prikeyenc []byte) error {

	jsondata,err := cs.readDB()
	if err != nil {
		return err
	}
	//fmt.Printf("%s\n",jsondata)

	err = json.Unmarshal(jsondata, &cs.fd)
	if err != nil {
		fmt.Println("error:", err)
	}

	cs.fd.PublicKeyHash = fmt.Sprintf("%x",pubkeyhash)
	cs.fd.PrivateKeyEncrypted = fmt.Sprintf("%x",prikeyenc)

	jsonblob, err := json.Marshal(cs.fd)
	if err != nil {
		fmt.Println("error:", err)
	}

	cs.writeDB( jsonblob )

	/*
		err := cs.OpenDB(cs.path)
		defer cs.CloseDB()
		if err != nil {
			return err
		}
	*/
	/*
		// insert
		stmt, err := cs.db.Prepare("INSERT INTO Account(PublicKeyHash,PrivateKeyEncrypted) values(?,?)")
		if err != nil {
			return err
		}

		res, err := stmt.Exec(pubkeyhash, prikeyenc)
		if err != nil {
			return err
		}

		_, err = res.LastInsertId()
		if err != nil {
			return err
		}

		//fmt.Printf("INSERT Account: [%x] %x\n",  pubkeyhash, prikeyenc )
	*/
	return nil
}

func (cs * FileStore) LoadAccountData(index int) ([]byte,[]byte,error) {
	jsondata,err := cs.readDB()
	if err != nil {
		return nil,nil,err
	}

	err = json.Unmarshal(jsondata, &cs.fd)
	if err != nil {
		fmt.Println("error:", err)
	}

	publickeyHash,err := hex.DecodeString(cs.fd.PublicKeyHash)
	privatekeyEncrypted,err := hex.DecodeString(cs.fd.PrivateKeyEncrypted)

	return publickeyHash,privatekeyEncrypted,err

/*
	err := cs.OpenDB(cs.path)
	defer cs.CloseDB()
	if err != nil {
		return nil,nil,err
	}

	rows, _ := cs.db.Query("SELECT * FROM Account")
	i := 0
	for rows.Next() {
		var pubkeyhash []byte
		var prikeyenc []byte

		rows.Scan(&pubkeyhash, &prikeyenc)

		if index == i {
			return pubkeyhash, prikeyenc, nil
		}

		i ++
	}
*/
	//return nil,nil,NewDetailErr(errors.New("[LoadAccountData] Index not found"), ErrNoCode, "")
}

func (cs * FileStore) LoadContractData( index int ) ([]byte,[]byte,[]byte,error) {

	jsondata,err := cs.readDB()
	if err != nil {
		return nil,nil,nil,err
	}

	err = json.Unmarshal(jsondata, &cs.fd)
	if err != nil {
		fmt.Println("error:", err)
	}

	scriptHash,err := hex.DecodeString(cs.fd.ScriptHash)
	publickeyHash,err := hex.DecodeString(cs.fd.PublicKeyHash)
	rawData,err := hex.DecodeString(cs.fd.RawData)

	return scriptHash,publickeyHash,rawData,err

/*
	err := cs.OpenDB(cs.path)
	defer cs.CloseDB()
	if err != nil {
		return nil,nil,nil,err
	}

	rows, _ := cs.db.Query("SELECT * FROM Contract")
	i := 0
	for rows.Next() {
		var sh []byte
		var ph []byte
		var rd []byte

		rows.Scan(&sh, &ph, &rd)

		if index == i {
			return sh, ph, rd, nil
		}

		i ++
	}
*/
	return nil,nil,nil,NewDetailErr(errors.New("[LoadContractData] Index not found"), ErrNoCode, "")
}

func (cs * FileStore) SaveContractData( ct * ct.Contract) error {

	jsondata,err := cs.readDB()
	if err != nil {
		return err
	}
	//fmt.Printf("%s\n",jsondata)

	err = json.Unmarshal(jsondata, &cs.fd)
	if err != nil {
		fmt.Println("error:", err)
	}

	cs.fd.ScriptHash = fmt.Sprintf("%x",ct.ProgramHash.ToArray())
	cs.fd.PublicKeyHash = fmt.Sprintf("%x",ct.OwnerPubkeyHash.ToArray())
	cs.fd.RawData = fmt.Sprintf("%x",ct.ToArray())

	jsonblob, err := json.Marshal(cs.fd)
	if err != nil {
		fmt.Println("error:", err)
	}

	cs.writeDB( jsonblob )

	//fmt.Printf("[SaveContractData]->[Contract] ScriptHash: %s\n", cs.fd.ScriptHash)
	//fmt.Printf("[SaveContractData]->[Contract] PublicKeyHash: %s\n", cs.fd.PublicKeyHash)
	//fmt.Printf("[SaveContractData]->[Contract] RawData: %s\n", cs.fd.RawData)
	//fmt.Printf("[SaveContractData] Code: %x\n", ct.Code)
	//fmt.Printf("[SaveContractData] Parameters: %x\n", ct.Parameters)

/*
	err := cs.OpenDB(cs.path)
	defer cs.CloseDB()
	if err != nil {
		return err
	}

	// insert Address
	stmt, err := cs.db.Prepare("INSERT INTO Address(ScriptHash) values(?)")
	if err != nil {
		return err
	}

	sh := ct.ProgramHash.ToArray()

	res, err := stmt.Exec(sh)
	if err != nil {
		return err
	}

	_, err = res.LastInsertId()
	if err != nil {
		return err
	}

	//fmt.Printf("[SaveContractData]->[Address] ScriptHash: %x\n", sh)

	// insert Contract
	stmt, err = cs.db.Prepare("INSERT INTO Contract(ScriptHash,PublicKeyHash,RawData) values(?,?,?)")
	if err != nil {
		return err
	}

	//sh := ct.ProgramHash.ToArray()
	ph := ct.OwnerPubkeyHash.ToArray()
	rd := ct.ToArray()

	res, err = stmt.Exec(sh, ph, rd )
	if err != nil {
		return err
	}

	_, err = res.LastInsertId()
	if err != nil {
		return err
	}
*/
	//fmt.Printf("[SaveContractData]->[Contract] ScriptHash: %x\n", sh)
	//fmt.Printf("[SaveContractData]->[Contract] PublicKeyHash: %x\n", ph)
	//fmt.Printf("[SaveContractData]->[Contract] RawData: %x\n", rd)
	//fmt.Printf("[SaveContractData] Code: %x\n", ct.Code)
	//fmt.Printf("[SaveContractData] Parameters: %x\n", ct.Parameters)

	return nil
}