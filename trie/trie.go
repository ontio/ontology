package trie

import (
	"bytes"
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"fmt"
	"reflect"
)

type DatabaseReader interface {
	Get(key []byte) ([]byte, error)
	Has(key []byte) (bool, error)
}

type DatabaseWriter interface {
	Put(key, value []byte) error
}

type Database interface {
	DatabaseReader
	DatabaseWriter
}

type Trie struct {
	root     node
	db       Database
	rootHash common.Uint256
}

func New(hash common.Uint256, db Database) (*Trie, error) {
	trie := &Trie{db: db, rootHash: hash}
	if (hash != common.Uint256{}) && db != nil {
		root, err := trie.resolveHash(hash.ToArray(), nil)
		if err != nil {
			return nil, err
		}
		trie.root = root
	}
	return trie, nil
}

func (t *Trie) Get(key []byte) []byte {
	res, err := t.TryGet(key)
	if err != nil {
		log.Error(fmt.Sprintf("Unhandled trie error: %v", err))
		return nil
	}
	return res
}

func (t *Trie) TryGet(key []byte) ([]byte, error) {
	key = keyBytesToHex(key)
	value, newRoot, err := t.tryGet(t.root, key, 0)
	if err != nil {
		return nil, err
	}
	t.root = newRoot
	return value, nil
}

func (t *Trie) tryGet(origNode node, key []byte, pos int) (value []byte, newNode node, err error) {
	switch n := origNode.(type) {
	case nil:
		return nil, nil, nil
	case valueNode:
		return n, n, nil
	case *shortNode:
		if len(key) - pos < len(n.Key) || !bytes.Equal(n.Key, key[pos:pos + len(n.Key)]) {
			return nil, n, nil
		}
		value, newNode, err = t.tryGet(n.Val, key, pos + len(n.Key))
		if err != nil {
			return nil, n, err
		}
		n = n.copy()
		n.Val = newNode
		return value, n, nil
	case *fullNode:
		value, newNode, err = t.tryGet(n.Children[key[pos]], key, pos + 1)
		if err != nil {
			return nil, n, err
		}
		n = n.copy()
		n.Children[key[pos]] = newNode
		return value, n, nil
	case hashNode:
		child, err := t.resolveHash(n, key[:pos])
		if err != nil {
			return nil, n, err
		}
		return t.tryGet(child, key, pos)
	default:
		panic(fmt.Sprintf("invalid find node type: %v", origNode))
	}
}

func (t *Trie) Update(key, value []byte) {
	if err := t.TryUpdate(key, value); err != nil {
		log.Error(fmt.Sprintf("Unhandled trie error: %v", err))
	}
}

func (t *Trie) TryUpdate(key, value [] byte) error {
	k := keyBytesToHex(key)
	n, err := t.insert(t.root, nil, k, valueNode(value))
	if err != nil {
		return err
	}
	t.root = n
	return nil
}

func (t *Trie) insert(n node, prefix, key []byte, value node) (node, error) {
	if len(key) == 0 {
		return value, nil
	}
	switch n := n.(type) {
	case *shortNode:
		matchLen := prefixLen(key, n.Key)
		if matchLen == len(n.Key) {
			nn, err := t.insert(n.Val, append(prefix, key[:matchLen]...), key[matchLen:], value)
			if err != nil {
				return nil, err
			}
			return &shortNode{Key: n.Key, Val: nn}, nil
		}
		branch := &fullNode{flags: nodeFlag{}}
		var err error
		branch.Children[n.Key[matchLen]], err = t.insert(nil, append(prefix, n.Key[:matchLen + 1]...), n.Key[matchLen + 1:], n.Val)
		if err != nil {
			return nil, err
		}
		branch.Children[key[matchLen]], err = t.insert(nil, append(prefix, key[:matchLen + 1]...), key[matchLen + 1:], value)
		if err != nil {
			return nil, err
		}
		if matchLen == 0 {
			return branch, nil
		}
		return &shortNode{key[:matchLen], branch, nodeFlag{}}, nil
	case *fullNode:
		nn, err := t.insert(n.Children[key[0]], append(prefix, key[0]), key[1:], value)
		if err != nil {
			return nil, err
		}
		n = n.copy()
		n.Children[key[0]] = nn
		return n, nil
	case nil:
		return &shortNode{Key: key, Val: value}, nil
	case hashNode:
		rn, err := t.resolveHash(n, prefix)
		if err != nil {
			return nil, err
		}
		nn, err := t.insert(rn, prefix, key, value)
		if err != nil {
			return rn, err
		}
		return nn, nil
	default:
		panic(fmt.Sprintf("invalid insert node type : %v", reflect.TypeOf(n)))
	}
}

func (t *Trie) Delete(key [] byte) {
	if err := t.TryDelete(key); err != nil {
		log.Error(fmt.Sprintf("Unhandled trie error: %v", err))
	}
}

func (t *Trie) TryDelete(key []byte) error {
	k := keyBytesToHex(key)
	n, err := t.delete(t.root, nil, k)
	if err != nil {
		return err
	}
	t.root = n
	return nil
}

func (t *Trie) delete(n node, prefix, key []byte) (node, error) {
	switch n := n.(type) {
	case *shortNode:
		matchLen := prefixLen(key, n.Key)
		if matchLen < len(n.Key) {
			return n, nil
		}
		if matchLen == len(key) {
			return nil, nil
		}
		child, err := t.delete(n.Val, append(prefix, key[:len(n.Key)]...), key[len(n.Key):])
		if err != nil {
			return n, err
		}
		switch child := child.(type) {
		case *shortNode:
			return &shortNode{Key: concat(n.Key, child.Key...), Val: child.Val}, nil
		default:
			return &shortNode{Key: n.Key, Val: child}, nil
		}
	case *fullNode:
		nn, err := t.delete(n.Children[key[0]], append(prefix, key[0]), key[1:])
		if err != nil {
			return n, err
		}
		n = n.copy()
		n.flags = nodeFlag{}
		n.Children[key[0]] = nn
		pos := -1
		for i, ch := range n.Children {
			if ch != nil {
				if pos == -1 {
					pos = i
				} else {
					pos = -2
					break
				}
			}
		}
		if pos >= 0 {
			if pos != 16 {
				cNode, err := t.resolve(n.Children[pos], prefix)
				if err != nil {
					return nil, err
				}
				if c, ok := cNode.(*shortNode); ok {
					k := append([]byte{byte(pos)}, c.Key...)
					return &shortNode{Key: k, Val: c.Val}, nil
				}
			}
			return &shortNode{Key: []byte{byte(pos)}, Val: n.Children[pos]}, nil
		}
		return n, nil
	case valueNode:
		return nil, nil
	case nil:
		return nil, nil
	case hashNode:
		rn, err := t.resolveHash(n, prefix)
		if err != nil {
			return nil, err
		}
		nn, err := t.delete(rn, prefix, key)
		if err != nil {
			return rn, err
		}
		return nn, nil
	default:
		panic(fmt.Sprintf("Invalid Delete Node Type: %v", n))
	}
}

func (t *Trie) Commit() (common.Uint256, error) {
	if t.db == nil {
		panic("Commit data to trie whit nil database")
	}
	return t.commitTo(t.db)
}

func (t *Trie) commitTo(db DatabaseWriter) (common.Uint256, error) {
	hash, cached, err := t.hashRoot(db)
	if err != nil {
		return common.Uint256{}, err
	}
	t.root = cached
	u160, _ := common.Uint256ParseFromBytes(hash.(hashNode))
	return u160, nil
}

func (t *Trie) Hash() common.Uint256 {
	hash, cached, _ := t.hashRoot(nil)
	t.root = cached
	u, _ := common.Uint256ParseFromBytes(hash.(hashNode))
	return u
}

func (t *Trie) hashRoot(db DatabaseWriter) (node, node, error) {
	if t.root == nil {
		return hashNode(nil), nil, nil
	}
	h := newHasher()
	defer returnHasherToPool(h)
	return h.hash(t.root, db, true)
}

func (t *Trie) resolve(n node, prefix []byte) (node, error) {
	if n, ok := n.(hashNode); ok {
		return t.resolveHash(n, prefix)
	}
	return n, nil
}

func (t *Trie) resolveHash(n hashNode, prefix []byte) (node, error) {
	enc, err := t.db.Get(n)
	if err != nil {
		return nil, err
	}
	dec := mustDecodeNode(n, enc)
	return dec, nil
}

func concat(s1 []byte, s2 ...byte) []byte {
	r := make([]byte, len(s1) + len(s2))
	copy(r, s1)
	copy(r[len(s1):], s2)
	return r
}




