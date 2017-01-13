package vm
import (
	"testing"
	"strings"
	"bytes"
	"fmt"
)

func TestIndex(t *testing.T) {
	const s, sep, want = "chicken", "ken", 4
	got := strings.Index(s, sep)
	if got != want {
		t.Errorf("Index(%q,%q) = %v; want %v", s, sep, got, want)// 注意原slide中 的got和want写反了
	}
}

func ExampleHello(){
	fmt.Println("Hello world!")
	// Output: Hello world!
}

func ExampleLen(){
	var vr * VmReader
	str := "hello"
	buf := bytes.NewBufferString(str)
	b := buf.Bytes()

	vr = NewVmReader( b )
	fmt.Println( vr.ReadVarString() )
	// Output: hello
}
