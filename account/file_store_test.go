package account

import (
	"testing"
	"os"
)

func TestSaveStoredData(t *testing.T){
	name := "IV"
	value := []byte("123456")
	path := "./fs.dat"
	fileStore := &FileStore{path:path}
	fileStore.BuildDatabase(path)

	err := fileStore.SaveStoredData(name, value)
	if err != nil {
		t.Errorf("SaveStoredData error:%s")
		return
	}

	v, err := fileStore.LoadStoredData(name)
	if err != nil {
		 t.Errorf("LoadStoredData error:%s", err)
		 return
	}

	if string(v) != string(value){
		t.Errorf("LoadStoredData value:%x != %x", v,value)
		return
	}
	os.RemoveAll(path)
	os.RemoveAll("ActorLog")
 }

func TestSaveAccountStore(t *testing.T){
	path := "./fs.dat"
	fileStore := &FileStore{path:path}
	fileStore.BuildDatabase(path)

	psd := []byte("123456")
	pkd := []byte("123456")
	err := fileStore.SaveAccountData(pkd, psd)
	if err != nil {
		t.Errorf("SaveAccountData error:%s", err)
		return
	}

	pkd1, psd1 , err := fileStore.LoadAccountData(0)
	if err != nil {
		t.Errorf("LoadAccountData error:%s", err)
		return
	}

	if string(pkd1) != string(pkd) {
		t.Errorf("TestSaveAccountStore Pk:%x != %x", pkd1, pkd)
		return
	}

	if string(psd1) != string(psd) {
		t.Errorf("TestSaveAccountStore PS:%x != %x", psd1, psd)
		return
	}

	os.RemoveAll(path)
	os.RemoveAll("ActorLog")
}
