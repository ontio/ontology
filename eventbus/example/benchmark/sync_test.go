/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

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
