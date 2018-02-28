package benmarks

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/Ontology/eventbus/actor"
)

type ping struct{ val int }

func BenchmarkSyncTest(b *testing.B) {
	defer time.Sleep(10 * time.Microsecond)
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer runtime.GOMAXPROCS(1)
	b.ReportAllocs()
	b.ResetTimer()
	props := actor.FromFunc(func(context actor.Context) {
		switch msg := context.Message().(type) {
		case *ping:
			val := msg.val
			context.Sender().Tell(&ping{val: val + 1})
		}
	})
	actora := actor.Spawn(props)
	iterations := int64(b.N)
	for i := int64(0); i < iterations; i++ {
		value := actora.RequestFuture(&ping{val: 1}, 50*time.Millisecond)
		res, err := value.Result()
		if err != nil {
			fmt.Printf("sync send msg error,%s,%d", err, res)
		}
	}
	b.StopTimer()
}
