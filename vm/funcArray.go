package vm

import(
	"sort"
	"math/big"
)

func opArraySize(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	arr :=  e.Stack.Pop()
	e.Stack.Push(NewStackItem(arr.Count()))

	return NONE,nil
}

func opPack(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	c :=  int(e.Stack.Pop().ToBigInt().Int64())
	if e.Stack.Count() < c {return FAULT,nil}
	arr := []StackItem{}

	for{
		if(c > 0) {arr = append(arr, *e.Stack.Pop())}
		c--
	}
	e.Stack.Push(NewStackItem(arr))

	return NONE,nil
}

func opUnpack(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	arr :=  e.Stack.Pop().GetArray()
	for _,si := range arr {
		e.Stack.Push(NewStackItem((si)))
	}
	e.Stack.Push(NewStackItem(len(arr)))

	return NONE,nil
}

func opDistinct(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	e.Stack.Push(e.Stack.Pop().Distinct())

	return NONE,nil
}

func opSort(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	var biSorter BigIntSorter
	biSorter = e.Stack.Pop().GetIntArray()

	sort.Sort(biSorter)
	e.Stack.Push(NewStackItem((biSorter)))

	return NONE,nil
}

func opReverse(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	arr := e.Stack.Pop().Reverse()
	e.Stack.Push(&arr)

	return NONE,nil
}

func opConcat(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	c :=  int(e.Stack.Pop().ToBigInt().Int64())
	if c == 1 {return FAULT,nil}
	if e.Stack.Count() < c {return FAULT,nil}
	item := e.Stack.Pop()
	c--
	for {
		c--
		if(c>0){
			item =  e.Stack.Pop().Concat(item)
		}
	}

	e.Stack.Push(item)

	return NONE,nil
}

func opUnion(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	c :=  int(e.Stack.Pop().ToBigInt().Int64())
	if c == 1 {return FAULT,nil}
	if e.Stack.Count() < c {return FAULT,nil}
	item := e.Stack.Pop()
	c--
	for {
		c--
		if(c>0){
			item =  e.Stack.Pop().Concat(item)
		}
	}

	e.Stack.Push(item.Distinct())

	return NONE,nil
}

func opIntersect(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	c :=  int(e.Stack.Pop().ToBigInt().Int64())
	if c == 1 {return FAULT,nil}
	if e.Stack.Count() < c {return FAULT,nil}
	item := e.Stack.Pop()
	c--
	for {
		c--
		if(c>0){
			item =  e.Stack.Pop().Intersect(item)
		}
	}

	e.Stack.Push(item)

	return NONE,nil
}

func opExcept(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	e.Stack.Push(x1.Except(x2))

	return NONE,nil
}

func opTake(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	count :=  int(e.Stack.Pop().ToBigInt().Int64())
	e.Stack.Push(e.Stack.Pop().Take(count))

	return NONE,nil
}

func opSkip(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	count :=  int(e.Stack.Pop().ToBigInt().Int64())
	e.Stack.Push(e.Stack.Pop().Take(count))

	return NONE,nil
}

func opPickItem(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	count :=  int(e.Stack.Pop().ToBigInt().Int64())
	e.Stack.Push(e.Stack.Pop().ElementAt(count))

	return NONE,nil
}

func opAll(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	bs := e.Stack.Pop().GetBoolArray()
	all := true
	for _,b := range bs {
		if !b {all = false; break}
	}
	e.Stack.Push(NewStackItem(all))

	return NONE,nil
}

func opAny(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	bs := e.Stack.Pop().GetBoolArray()
	any := false
	for _,b := range bs {
		if b {any = true; break}
	}
	e.Stack.Push(NewStackItem(any))

	return NONE,nil
}

func opSum(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	is := e.Stack.Pop().GetIntArray()
	sum := SumBigInt(is)
	e.Stack.Push(NewStackItem(sum))

	return NONE,nil
}

func opAverage(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	arr := e.Stack.Pop()
	arrCount := arr.Count()
	if arrCount < 1 {return FAULT,nil}
	is := e.Stack.Pop().GetIntArray()
	sum := SumBigInt(is)
	avg := sum.Div(&sum,big.NewInt(int64(arrCount)))
	e.Stack.Push(NewStackItem(*avg))

	return NONE,nil
}

func opMaxItem(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return  FAULT,nil}
	e.Stack.Push(NewStackItem(MinBigInt(e.Stack.Pop().GetIntArray())))

	return NONE,nil
}

func opMinItem(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return  FAULT,nil}
	e.Stack.Push(NewStackItem(MinBigInt(e.Stack.Pop().GetIntArray())))

	return NONE,nil
}