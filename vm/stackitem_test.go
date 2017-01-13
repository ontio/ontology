package vm
import (
	"fmt"
)

func PrintStack( info string, stackItem  StackItem ) {
	var length = len(stackItem.array)

	fmt.Print( info )
	fmt.Print( " " )

	for i := 0; i < length; i++ {
		fmt.Print( stackItem.array[i] )
		fmt.Print( " " )
	}
	fmt.Println()
}

func ExampleStackItem(){

	var stackItem *StackItem
	var stackItem2 *StackItem
	var stackItem3 *StackItem
	var stackItem4 *StackItem
	var stackItem5 *StackItem
	var stackItem6 *StackItem
	var stackItem7 *StackItem
	var stackItem8 StackItem
	var stackItem9 []StackItem

	bb1 := []byte{1,2,3}
	stackItem = NewStackItem(bb1)
	//fmt.Println(stackItem.Count());

	bb2 := []byte{4,5,6}
	stackItem.Concat( NewStackItem(bb2) )
	//fmt.Println(stackItem.Count());

	bb3 := []byte{7,8,9}
	stackItem.Concat( NewStackItem(bb3) )
	fmt.Println( "Count() test:", stackItem.Count());
	PrintStack( "Concat() test:", *stackItem )

	stackItem2 = NewStackItem(bb2)
	PrintStack( "NewStackItem() test:", *stackItem2 )

	stackItem3 = stackItem.Except( stackItem2 )
	PrintStack( "Except() test:", *stackItem3 )

	stackItem4 = stackItem.Intersect( stackItem2 )
	PrintStack( "Intersect() test:", *stackItem4 )

	stackItem5 = stackItem.Take(3)
	PrintStack( "Take() test:", *stackItem5 )

	stackItem6 = stackItem.Skip(0)
	PrintStack( "Skip() test:", *stackItem6 )

	stackItem7 = stackItem.ElementAt(1)
	if ( stackItem7 != nil ){
		PrintStack( "ElementAt() test:", *stackItem7 )
	}

	stackItem8 = stackItem.Reverse()
	PrintStack( "Reverse() test:", stackItem8 )

	stackItem9 = stackItem.GetArray()
	PrintStack( "GetArray() test:", stackItem9[0] )
	PrintStack( "GetArray() test:", stackItem9[1] )
	PrintStack( "GetArray() test:", stackItem9[2] )

	bb4 := stackItem.GetBytes()
	fmt.Println( "GetBytes() test:", bb4 )

	bb5 := stackItem.GetBytesArray()
	fmt.Println( "GetBytesArray() test:", bb5[0] )
	fmt.Println( "GetBytesArray() test:", bb5[1] )
	fmt.Println( "GetBytesArray() test:", bb5[2] )

	bb6 := stackItem.GetIntArray()
	fmt.Println( "GetIntArray() test:", bb6[0] )
	fmt.Println( "GetIntArray() test:", bb6[1] )
	fmt.Println( "GetIntArray() test:", bb6[2] )

	bb7 := stackItem.GetBoolArray()
	fmt.Println( "GetBoolArray() test:", bb7[0] )
	fmt.Println( "GetBoolArray() test:", bb7[1] )
	fmt.Println( "GetBoolArray() test:", bb7[2] )

	b  := stackItem.ToBool()
	fmt.Println( "ToBool() test:", b )

	bi := stackItem.ToBigInt()
	fmt.Println( "ToBigInt() test:", bi )

	/*
	output: ok
	*/
}
