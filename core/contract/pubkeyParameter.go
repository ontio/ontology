package contract

type PubkeyParameter struct {
	PubKey string
	Parameter string
}



type ParameterIndex struct {
	Parameter []byte
	Index int
}

type ParameterIndexSlice []ParameterIndex

func (p ParameterIndexSlice) Len() int           { return len(p) }
func (p ParameterIndexSlice) Less(i, j int) bool { return p[i].Index < p[j].Index }
func (p ParameterIndexSlice) Swap(i, j int)      { p[i].Index, p[j].Index = p[j].Index, p[i].Index }