package trie

import (
	"fmt"
	"io"
	"github.com/Ontology/rlp"
	"github.com/Ontology/common"
)

var indices = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "17"}

type node interface {
	fString(string) string
	cache() (hashNode, bool)
}

type (
	fullNode struct {
		Children [17]node // Actual trie node data to encode/decode (needs custom encoder)
		flags    nodeFlag
	}
	shortNode struct {
		Key   []byte
		Val   node
		flags nodeFlag
	}
	hashNode  []byte
	valueNode []byte
)

func (n  *fullNode) copy() *fullNode {
	c := *n; return &c
}
func (n *shortNode) copy() *shortNode {
	c := *n; return &c
}

func (n *fullNode) cache() (hashNode, bool) {
	return n.flags.hash, n.flags.dirty
}
func (n *shortNode) cache() (hashNode, bool) {
	return n.flags.hash, n.flags.dirty
}
func (n hashNode) cache() (hashNode, bool) {
	return nil, true
}
func (n valueNode) cache() (hashNode, bool) {
	return nil, true
}

func (n *fullNode) fString(ind string) string {
	resp := fmt.Sprintf("[\n%s  ", ind)
	for i, node := range n.Children {
		if node == nil {
			resp += fmt.Sprintf("%s: <nil> ", indices[i])
		} else {
			resp += fmt.Sprintf("%s: %v", indices[i], node.fString(ind + "  "))
		}
	}
	return resp + fmt.Sprintf("\n%s] ", ind)
}

func (n *shortNode) fString(ind string) string {
	return fmt.Sprintf("{%v: %v} ", n.Key, n.Val.fString(ind + "  "))
}
func (n hashNode) fString(ind string) string {
	return fmt.Sprintf("<%x>", []byte(n))
}
func (n valueNode) fString(ind string) string {
	return fmt.Sprintf("%s", string(n))
}

func (n *fullNode) String() string {
	return n.fString("")
}
func (n *shortNode) String() string {
	return n.fString("")
}
func (n hashNode) String() string {
	return n.fString("")
}
func (n valueNode) String() string {
	return n.fString("")
}

type nodeFlag struct {
	hash  hashNode
	dirty bool
}

func mustDecodeNode(hash, buf []byte) node {
	n, err := decodeNode(hash, buf)
	if err != nil {
		panic(fmt.Sprintf("node %x, %v", hash, err))
	}
	return n
}

func decodeNode(hash, buf []byte) (node, error) {
	if len(buf) == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	elems, _, err := rlp.SplitList(buf)
	if err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}
	switch c, _ := rlp.CountValues(elems); c{
	case 2:
		n, err := decodeShort(hash, elems)
		if err != nil {
			return nil, err
		}
		return n, nil
	case 17:
		n, err := decodeFull(hash, elems)
		if err != nil {
			return nil, err
		}
		return n, nil
	default:
		return nil, fmt.Errorf("invalid number of list elements: %v", c)
	}
}

func decodeShort(hash, els []byte) (*shortNode, error) {
	kBuf, rest, err := rlp.SplitString(els)
	if err != nil {
		return nil, err
	}
	flag := nodeFlag{hash: hash}
	key := compactToHex(kBuf)
	if hasTerm(key) {
		val, _, err := rlp.SplitString(rest)
		if err != nil {
			return nil, err
		}
		return &shortNode{key, append(valueNode{}, val...), flag}, nil
	}
	r, _, err := decodeRef(rest)
	if err != nil {
		return nil, err
	}
	return &shortNode{key, r, flag}, nil
}

func decodeFull(hash, els []byte) (*fullNode, error) {
	n := &fullNode{flags: nodeFlag{hash: hash}}
	for i := 0; i < 16; i++ {
		cld, rest, err := decodeRef(els)
		if err != nil {
			return n, err
		}
		n.Children[i], els = cld, rest
	}
	val, _, err := rlp.SplitString(els)
	if err != nil {
		return n, err
	}
	if len(val) > 0 {
		n.Children[16] = append(valueNode{}, val...)
	}
	return n, nil
}

const hashLen = len(common.Uint256{})

func decodeRef(buf []byte) (node, []byte, error) {
	kind, val, rest, err := rlp.Split(buf)
	if err != nil {
		return nil, buf, err
	}
	switch {
	case kind == rlp.List:
		if size := len(buf) - len(rest); size > hashLen {
			return nil, buf, fmt.Errorf("oversized embedded node (size is %d bytes, want size < %d)", size, hashLen)
		}
		n, err := decodeNode(nil, buf)
		return n, rest, err
	case kind == rlp.String && len(val) == 0:
		return nil, rest, nil
	case kind == rlp.String && len(val) == 32:
		return append(hashNode{}, val...), rest, nil
	default:
		return nil, nil, fmt.Errorf("[decodeRef] invalid RLP string size %d (want 0 or 32)", len(val))
	}
}

