package vm

import (
	"fmt"
)

func ExampleOpStack(){

	os := NewOpStack()
	fmt.Println( "NewOpStack() test:", os.Element )

	os.Push( NewStackItem([]byte{'A','D','D'}) )
	fmt.Println( "Push() test:", os.Element[0].array )
	os.Push( NewStackItem([]byte{'S','U','B'}) )
	fmt.Println( "Push() test:", os.Element[1].array )
	os.Push( NewStackItem([]byte{'M','U','L'}) )
	fmt.Println( "Push() test:", os.Element[2].array )

	fmt.Println( "Count() test:", os.Count() )

	si := os.Peek()
	if ( si != nil ){
		fmt.Println( "Peek() test:", si )
	}

	fmt.Println( "Pop() test:", os.Element )

	sip := os.Pop()
	if ( sip != nil ){
		fmt.Println( "Pop() test:", sip )
		fmt.Println( "Pop() test:", os.Element )
	}

	sip = os.Pop()
	if ( sip != nil ){
		fmt.Println( "Pop() test:", sip )
		fmt.Println( "Pop() test:", os.Element )
	}

	sip = os.Pop()
	if ( sip != nil ){
		fmt.Println( "Pop() test:", sip )
		fmt.Println( "Pop() test:", os.Element )
	}

	// output: ok
}