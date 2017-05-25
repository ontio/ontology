package utils

type RandomAccessStack struct {
	Element []interface{}
}

func NewRandAccessStack() *RandomAccessStack {
	var ras RandomAccessStack
	ras.Element = make([]interface{}, 0)
	return &ras
}

func (ras *RandomAccessStack) Count() int {
	return len(ras.Element)
}

func (ras *RandomAccessStack) Insert(index int, t interface{}) {
	l := len(ras.Element)
	if index > l {
		return
	}
	if index == 0 {
		ras.Element = append(ras.Element, t)
		return
	}

	var array = make([]interface{}, 0, l+1)
	index = l - index
	array = append(array, ras.Element[:index])
	array = append(array, t)
	array = append(array, ras.Element[index:]...)

	ras.Element = array
}

func (ras *RandomAccessStack) Peek(index int) interface{} {
	l := len(ras.Element)
	if index >= l {
		return nil
	}
	index = l - index
	return ras.Element[index-1]
}

func (ras *RandomAccessStack) Remove(index int) interface{} {
	l := len(ras.Element)
	if index >= l {
		return nil
	}
	index = l - index
	e := ras.Element[index-1]
	var si []interface{}
	si = append(ras.Element[:index-1], ras.Element[index:]...)
	ras.Element = si
	return e
}

func (ras *RandomAccessStack) Set(index int, t interface{}) {
	l := len(ras.Element)
	if index >= l {
		return
	}
	ras.Element[index] = t
}

func (ras *RandomAccessStack) Push(t interface{}) {
	ras.Insert(0, t)
}

func (ras *RandomAccessStack) Pop() interface{} {
	return ras.Remove(0)
}

func (ras *RandomAccessStack) Swap(i, j int) {
	ras.Element[i], ras.Element[j] = ras.Element[j], ras.Element[i]
}
