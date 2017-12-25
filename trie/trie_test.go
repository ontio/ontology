package trie

import (
	"testing"
	"bytes"
	"github.com/Ontology/common"
)

func newEmpty() *Trie {
	db := NewMemDatabase()
	trie, _ := New(common.Uint256{}, db)
	return trie
}

func TestNode(t *testing.T) {
	trie := newEmpty()

	trie.TryUpdate([]byte("123456"), []byte("asdfasdfasdfasdfasdfasdfasdfasdf"))
	trie.TryUpdate([]byte("12366"), []byte("wqeqweqweqweqweqwewerwerwerwerwerwerwerwerwerwerwerwerwerwqeqweqweqweqweqwewerwerwerwerwerwerwerwerwerwerwerwerwerwqeqweqweqweqweqwewerwerwerwerwerwerwerwerwerwerwerwerwerwqeqweqweqweqweqwewerwerwerwerwerwerwerwerwerwerwerwerwerwqeqweqweqweqweqwewerwerwerwerwerwerwerwerwerwerwerwerwerwqeqweqweqweqweqwewerwerwerwerwerwerwerwerwerwerwerwerwerwqeqweqweqweqweqwewerwerwerwerwerwerwerwerwerwerwerwerwer"))
	trie.TryUpdate([]byte("1234"), []byte("asdfasdfasdfasdfasdfasdfasdfasdf"))

	root, _ := trie.Commit()

	trie, _ = New(*root, trie.db)

	v, err := trie.TryGet([]byte("1234"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	t.Log("key 1234 value:", string(v))

	err = trie.TryUpdate([]byte("120099"), []byte("zxcvzxcvzxcvzxcvzxcvzxcvzxcvzxcv"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	v, err = trie.TryGet([]byte("120099"))
	if err != nil {
		t.Errorf("Wrong error: %v", err)
	}
	t.Log("key 120099 value:", string(v))

	err = trie.TryDelete([]byte("123456"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	v, err = trie.TryGet([]byte("123456"))
	if err != nil {
		t.Errorf("Wrong error: %v", err)
	}
}

func TestGet(t *testing.T) {
	trie := newEmpty()
	updateString(trie, "doe", "reindeer")
	updateString(trie, "dog", "puppy")
	updateString(trie, "dogglesworth", "cat")

	for i := 0; i < 2; i++ {
		res := getString(trie, "dog")
		if !bytes.Equal(res, []byte("puppy")) {
			t.Errorf("expected puppy got %x", res)
		}

		unknown := getString(trie, "unknown")
		if unknown != nil {
			t.Errorf("expected nil got %x", unknown)
		}

		if i == 1 {
			return
		}
		trie.Commit()
	}
}

func TestReplication(t *testing.T) {
	trie := newEmpty()
	vals := []struct{ k, v string }{
		{"do", "verb"},
		{"ether", "wookiedoo"},
		{"horse", "stallion"},
		{"shaman", "horse"},
		{"doge", "coin"},
		{"dog", "puppy"},
		{"somethingveryoddindeedthis is", "myothernodedata"},
	}
	for _, val := range vals {
		updateString(trie, val.k, val.v)
	}
	exp, err := trie.Commit()
	if err != nil {
		t.Fatalf("commit error: %v", err)
	}

	// create a new trie on top of the database and check that lookups work.
	trie2, err := New(*exp, trie.db)
	if err != nil {
		t.Fatalf("can't recreate trie at %x: %v", exp, err)
	}
	for _, kv := range vals {
		if string(getString(trie2, kv.k)) != kv.v {
			t.Errorf("trie2 doesn't have %q => %q", kv.k, kv.v)
		}
	}
	hash, err := trie2.Commit()
	if err != nil {
		t.Fatalf("commit error: %v", err)
	}
	if hash.CompareTo(*exp) != 0 {
		t.Errorf("root failure. expected %x got %x", exp, hash)
	}

	// perform some insertions on the new trie.
	vals2 := []struct{ k, v string }{
		{"do", "verb"},
		{"ether", "wookiedoo"},
		{"horse", "stallion"},
		{"shaman", "horse"},
		{"doge", "coin"},
		{"ether", ""},
		{"dog", "puppy"},
		{"somethingveryoddindeedthis is", "myothernodedata"},
		{"shaman", ""},
	}
	for _, val := range vals2 {
		updateString(trie2, val.k, val.v)
	}
	if hash := trie2.Hash(); hash.CompareTo(*exp) != 0 {
		t.Errorf("root failure. expected %x got %x", exp, hash)
	}
}

func TestEmptyValues(t *testing.T) {
	trie := newEmpty()

	vals := []struct{ k, v string }{
		{"do", "verb"},
		{"ether", "wookiedoo"},
		{"horse", "stallion"},
		{"shaman", "horse"},
		{"doge", "coin"},
		{"ether", ""},
		{"dog", "puppy"},
		{"shaman", ""},
	}
	for _, val := range vals {
		updateString(trie, val.k, val.v)
	}

	t.Log("shaman", getString(trie, vals[7].k))
}

type countingDB struct {
	Database
	gets map[string]int
}

func (db *countingDB) Get(key []byte) ([]byte, error) {
	db.gets[string(key)]++
	return db.Database.Get(key)
}

func TestCacheUnload(t *testing.T) {
	// Create test trie with two branches.
	trie := newEmpty()
	key1 := "---------------------------------"
	key2 := "---some other branch"
	updateString(trie, key1, "this is the branch of key1.")
	updateString(trie, key2, "this is the branch of key2.")
	root, _ := trie.Commit()

	// Commit the trie repeatedly and access key1.
	// The branch containing it is loaded from DB exactly two times:
	// in the 0th and 6th iteration.
	db := &countingDB{Database: trie.db, gets: make(map[string]int)}
	trie, _ = New(*root, db)
	//trie.SetCacheLimit(5)
	for i := 0; i < 12; i++ {
		getString(trie, key1)
		trie.Commit()
	}

	// Check that it got loaded two times.
	for dbkey, count := range db.gets {
		if count != 2 {
			t.Errorf("db key %x loaded %d times, want %d times", []byte(dbkey), count, 2)
		}
	}
}


//func TestPerform(t *testing.T) {
//	start := time.Now().Unix()
//	trie := newEmpty()
//	for i := 0; i < 100000; i++ {
//		start1 := time.Now().Unix()
//		trie, _ = New(trie.rootHash, trie.db)
//		for j := 0; j < 100000; j ++ {
//			updateString(trie, fmt.Sprintf("%d", i * j), fmt.Sprintf("%d", i *j))
//		}
//		trie.Commit()
//		end1 := time.Now().Unix()
//		fmt.Println(i, " spend time:", end1 - start1)
//	}
//	end := time.Now().Unix()
//	fmt.Println("spend time:", end - start)
//}

func getString(trie *Trie, k string) []byte {
	return trie.Get([]byte(k))
}

func updateString(trie *Trie, k, v string) {
	trie.Update([]byte(k), []byte(v))
}