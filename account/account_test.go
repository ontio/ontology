package account

import (
	"DNA/crypto"
	"fmt"
	"os"
	"path"
	"testing"
)

func TestClient(t *testing.T) {
	t.Log("created client start!")
	crypto.SetAlg(crypto.P256R1)
	dir := "./data/"
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		t.Log("create dir ", dir, " error: ", err)
	} else {
		t.Log("create dir ", dir, " success!")
	}
	for i := 0; i < 10000; i++ {
		p := path.Join(dir, fmt.Sprintf("wallet%d.txt", i))
		fmt.Println("client path", p)
		CreateClient(p, []byte(DefaultPin))
	}
}
