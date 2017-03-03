package client

import (
	ct "GoOnchain/core/contract"
	//. "GoOnchain/common"
	"fmt"
	"database/sql"
	"github.com/mattn/go-sqlite3"
	"os"
	. "GoOnchain/errors"
	"errors"
	"strings"
)

type ClientStore struct {
	db * sql.DB
	path string
}

func (cs * ClientStore) OpenDB(path string) error {
	cs.CloseDB()

	var err error
	cs.db, err = sql.Open("sqlite3", path)
	if err != nil {
		return err
	}

	return nil
}

func (cs * ClientStore) CloseDB() {
	if  cs.db != nil {
		cs.db.Close()
		cs.db = nil
	}
}

func (cs * ClientStore) BuildDatabase(path string) {
	sqlite3.Version()

	err := os.Remove(path)
	if err != nil {
		fmt.Println( err )
		//return
	}

	err = cs.OpenDB(path)
	defer cs.CloseDB()
	if err != nil {
		// TODO: error
		fmt.Println( err )
	}

	sql := `CREATE TABLE Account (
		    PublicKeyHash       BINARY    NOT NULL
						  CONSTRAINT PK_Account PRIMARY KEY,
		    PrivateKeyEncrypted VARBINARY NOT NULL
		);`
	cs.db.Exec(sql)

	sql = `CREATE TABLE Address (
		    ScriptHash BINARY NOT NULL
				    CONSTRAINT PK_Address PRIMARY KEY
		);`
	cs.db.Exec(sql)

	sql = `CREATE TABLE Coin (
		    TxId       BINARY  NOT NULL,
		    [Index]    INTEGER NOT NULL,
		    AssetId    BINARY  NOT NULL,
		    ScriptHash BINARY  NOT NULL,
		    State      INTEGER NOT NULL,
		    Value      INTEGER NOT NULL,
		    CONSTRAINT PK_Coin PRIMARY KEY (
			TxId,
			[Index]
		    ),
		    CONSTRAINT FK_Coin_Address_ScriptHash FOREIGN KEY (
			ScriptHash
		    )
		    REFERENCES Address (ScriptHash) ON DELETE CASCADE
		);`
	cs.db.Exec(sql)

	sql = `CREATE TABLE Contract (
		    ScriptHash    BINARY    NOT NULL
					    CONSTRAINT PK_Contract PRIMARY KEY,
		    PublicKeyHash BINARY    NOT NULL,
		    RawData       VARBINARY NOT NULL,
		    CONSTRAINT FK_Contract_Account_PublicKeyHash FOREIGN KEY (
			PublicKeyHash
		    )
		    REFERENCES Account (PublicKeyHash) ON DELETE CASCADE,
		    CONSTRAINT FK_Contract_Address_ScriptHash FOREIGN KEY (
			ScriptHash
		    )
		    REFERENCES Address (ScriptHash) ON DELETE CASCADE
		);`
	cs.db.Exec(sql)

	sql = `CREATE TABLE [Key] (
		    Name  VARCHAR   NOT NULL
				    CONSTRAINT PK_Key PRIMARY KEY,
		    Value VARBINARY NOT NULL
		);`
	cs.db.Exec(sql)

	sql = `CREATE TABLE [Transaction] (
		    Hash    BINARY    NOT NULL
				      CONSTRAINT PK_Transaction PRIMARY KEY,
		    Height  INTEGER,
		    RawData VARBINARY NOT NULL,
		    Time    TEXT      NOT NULL,
		    Type    INTEGER   NOT NULL
		);`
	cs.db.Exec(sql)

	//cs.db = db
	//fmt.Println("create table success.")
}

func (cs * ClientStore) SaveStoredData(name string,value []byte) error {

	err := cs.OpenDB(cs.path)
	defer cs.CloseDB()
	if err != nil {
		// TODO: error
		return err
	}

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

	return nil
}

func (cs * ClientStore) LoadStoredData(name string) ([]byte,error) {

	err := cs.OpenDB(cs.path)
	defer cs.CloseDB()
	if err != nil {
		// TODO: error
		return nil,err
	}

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

	return nil, NewDetailErr(errors.New("Can't find the key: " + name), ErrNoCode, "")
}

func (cs * ClientStore) SaveAccountData(pubkeyhash []byte,prikeyenc []byte) error {

	err := cs.OpenDB(cs.path)
	defer cs.CloseDB()
	if err != nil {
		return err
	}

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

	return nil
}

func (cs * ClientStore) LoadAccountData(index int) ([]byte,[]byte,error) {

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

	return nil,nil,NewDetailErr(errors.New("[LoadAccountData] Index not found"), ErrNoCode, "")
}

func (cs * ClientStore) LoadContractData( index int ) ([]byte,[]byte,[]byte,error) {

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

	return nil,nil,nil,NewDetailErr(errors.New("[LoadContractData] Index not found"), ErrNoCode, "")
}

func (cs * ClientStore) SaveContractData( ct * ct.Contract) error {

	err := cs.OpenDB(cs.path)
	defer cs.CloseDB()
	if err != nil {
		return err
	}

	// insert
	stmt, err := cs.db.Prepare("INSERT INTO Contract(ScriptHash,PublicKeyHash,RawData) values(?,?,?)")
	if err != nil {
		return err
	}

	sh := ct.ProgramHash.ToArray()
	ph := ct.OwnerPubkeyHash.ToArray()
	rd := ct.ToArray()

	res, err := stmt.Exec(sh, ph, rd )
	if err != nil {
		return err
	}

	_, err = res.LastInsertId()
	if err != nil {
		return err
	}

	//fmt.Printf("INSERT Contract: %x %x %x\n",  sh, ph, rd )
	fmt.Printf("[SaveContractData] ScriptHash: %x\n", sh)
	fmt.Printf("[SaveContractData] PublicKeyHash: %x\n", ph)
	fmt.Printf("[SaveContractData] RawData: %x\n", rd)
	fmt.Printf("[SaveContractData] Code: %x\n", ct.Code)
	fmt.Printf("[SaveContractData] Parameters: %x\n", ct.Parameters)

	return nil
}