package overlaydb

import "fmt"

func (db *MemDB) DumpToDot() string {
	out := `digraph g {
				graph [
				rankdir = "LR"
				style="filled"
				];
				node [
				fontsize = "16"
				shape = "ellipse"
				style="filled"
				];`
	out += "\n"

	nodes := ""
	edges := ""

	node := 0
	for {
		h := db.nodeData[node+nHeight]
		if h > db.maxHeight {
			h = db.maxHeight
		}
		n := db.nodeData[node]
		m := n + db.nodeData[node+nKey]
		k := db.kvData[n:m]
		v := db.kvData[m : m+db.nodeData[node+nVal]]

		nodes += genNode(node, k, v, h)
		edges += genEdges(node, db.nodeData[node+nNext:node+nNext+h])
		node = db.nodeData[node+nNext]
		if node == 0 {
			break
		}
	}
	out += nodes
	out += edges
	out += "}"

	return out
}

func genEdges(node int, nextNodes []int) string {
	str := ""
	for i, next := range nextNodes {
		if next != 0 {
			str += fmt.Sprintf(`"n%d":f%d -> "n%d":f%d ;`, node, i+1, next, i+1)
			str += "\n"
		}
	}

	return str
}

func genNode(node int, k []byte, v []byte, h int) string {
	str := fmt.Sprintf(`"n%d" [
	shape = "record"
	label = "`, node)
	for i := h; i > 0; i-- {
		str += fmt.Sprintf("<f%d> %d | ", i, i)
	}
	str += fmt.Sprintf(`%s:%s"];`, k, v)
	str += "\n"

	return str
}
