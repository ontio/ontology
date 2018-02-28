package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/Ontology/eventbus/actor"
)

type Ball struct {
	val int
}

var start, end int64

//func Benchmark_Division1(b *testing.B){
func main() {
	fmt.Printf("test performance")
	runtime.GOMAXPROCS(4)
	times := 10000000
	props := actor.FromFunc(func(context actor.Context) {
		switch msg := context.Message().(type) {

		case Ball:
			val := msg.val
			if val < times {
				context.Sender().Request(Ball{val: val + 1}, context.Self())
			} else {
				end = time.Now().UnixNano()
				fmt.Printf("end at time %d\n", end)
			}
		default:
		}
	})
	playerA, _ := actor.SpawnNamed(props, "playerA")
	playerB, _ := actor.SpawnNamed(props, "playerB")
	start = time.Now().UnixNano()
	fmt.Println("start at time:", start)
	playerA.Request(Ball{val: 1}, playerB)
	time.Sleep(10000 * time.Millisecond)
	fmt.Printf("run time:%d     elapsed time:%d ms", times, (end-start)/1000000)
}
