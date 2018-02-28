# Actor模型的Go语言实现

[进程内通讯](#进程内通讯)

[跨节点通讯](#跨节点通讯)

# 关于Actor模型
Actor是计算机科学领域中的一个并行计算模型，它把actors当做通用的并行计算原语：一个actor对接收到的消息做出响应，进行本地决策，可以创建更多的actor，或者发送更多的消息；同时准备接收下一条消息。

在Actor理论中，一切都被认为是actor，这和面向对象语言里一切都被看成对象很类似。但包括面向对象语言在内的软件通常是顺序执行的，而Actor模型本质上则是并发的。

每个Actor都有一个(只有一个)Mailbox。Mailbox相当于是一个小型的队列，一旦Sender发送消息，就是将该消息入队到Mailbox中。入队的顺序按照消息发送的时间顺序。Mailbox有多种实现，默认为FIFO。但也可以根据优先级考虑出队顺序，实现算法则不相同。

![actor](/resources/actor.png)

## Actor VS Channel

![actor](/resources/actors.png)


![channel](/resources/channel.png)

## Actor模型的优点：

1. actor的mailbox容量是无限的，不会造成写入时的阻塞

2. 每个actor中所有消息共用一个mailbox(channel)。

3. actor并不关心消息的发送方（writer）,可以对各模块间的逻辑进行解耦合。

4. actor可以部署在不同节点上。

## Actor模型的缺点

1. 因为Actor被设计为异步模型，同步调用的性能不高。


## 创建Actors

Props为声明如何创建Actors提供了基础，下面的例子通过定义如何处理消息的函数声明定义了Actor Props:
```go
var props Props = actor.FromFunc(func(c Context) {
	// process messages
})
```

另外，可以创建一个结构体，通过定义一个Receive方法，实现了Actor的接口：

```go
type MyActor struct {}

func (a *MyActor) Receive(c Context) {
	// process messages
}

var props Props = actor.FromProducer(func() Actor { return &MyActor{} })
```

Spawn和SpawnNamed使用给定的props去创建Actor的运行实例。一旦启动Actor就开始准备处理发来的消息。用系统给定的唯一名称来启动actor，使用：
```go
pid := actor.Spawn(props)
```
结果返回唯一的PID。自己命名PID请使用 SpawnNamed 来启动Actor。

每次一个actor启动时，一个新的邮箱会被创建并关联PID。消息会发送到该邮箱然后actor来处理这些消息。

## 处理消息

Actor通过Receive方法来处理消息，此函数定义为：
```go
Receive(c actor.Context)
```
系统会保证该方法被同步调用，因此无需做另外的保护措施。

## 与其他actors通讯

PID是向actors发送消息的主要接口，PID.Tell方法用于向该PID异步的发送消息：
```go
pid.Tell("Hello World")
```
根据不同的业务需求，actors之间的通讯可以异步或者同步进行，不论任何时候，actors总是通过PID来进行通讯。

当使用PID.Request或者PID.RequestFuture来进行消息发送时，接受消息的actor将会通过Context.Sender方法来回应发送者，该方法返回发送者的PID。

同步通讯方面，actor使用Future来实现，actor再继续下一步之前会等待结果获取。

向actor发送消息并等待结果获取，请使用RequestFuture方法，该方法会返回一个Future：
```go
f := actor.RequestFuture(pid,"Hello", 50 * time.Millisecond)
res, err := f.Result() // waits for pid to reply */
```

# 进程内通讯
## 性能
### 异步调用

protoactor-go目前可以每秒在两个actor之间传递200万条消息，并且能够保证消息的顺序。

```text
/app/go/bin/go build -o "/tmp/Build performanceTest.go and rungo" /app/gopath/src/github.com/OntologyNetwork/protoactor-go/examples/examples/performanceTest.go
start at time: 1516953710985385134
end at time 1516953716291953904
run time:10000000     elapsed time:5306 ms
```

### 串行同步调用

protoactor-go目前可以在串行同步调用的情况下每秒在client和server间传递超过50万条消息！

```text
goos: linux
goarch: amd64
pkg: github.com/OntologyNetwork/protoactor-go/examples/benchmark
benchmark                  iter                 time/iter          bytes alloc          allocs
---------                  ----                ---------           -----------          ------
BenchmarkSyncTest-4   	 1000000	      1967 ns/op	     432 B/op	      13 allocs/op
testing: BenchmarkSyncTest-4 left GOMAXPROCS set to 1
BenchmarkSyncTest-4   	 1000000	      1987 ns/op	     432 B/op	      13 allocs/op
testing: BenchmarkSyncTest-4 left GOMAXPROCS set to 1
BenchmarkSyncTest-4   	 1000000	      1952 ns/op	     432 B/op	      13 allocs/op
testing: BenchmarkSyncTest-4 left GOMAXPROCS set to 1
BenchmarkSyncTest-4   	 1000000	      1975 ns/op	     432 B/op	      13 allocs/op
testing: BenchmarkSyncTest-4 left GOMAXPROCS set to 1
BenchmarkSyncTest-4   	 1000000	      1987 ns/op	     432 B/op	      13 allocs/op
testing: BenchmarkSyncTest-4 left GOMAXPROCS set to 1
PASS
ok  	github.com/OntologyNetwork/protoactor-go/examples/benmarks	10.984s

```

## Hello world

```go
type Hello struct{ Who string }
type HelloActor struct{}

func (state *HelloActor) Receive(context actor.Context) {
    switch msg := context.Message().(type) {
    case Hello:
        fmt.Printf("Hello %v\n", msg.Who)
    }
}

func main() {
    props := actor.FromProducer(func() actor.Actor { return &HelloActor{} })
    pid := actor.Spawn(props)
    pid.Tell(Hello{Who: "Roger"})
    console.ReadLine()
}
```


## Two actors communicates each other

本例主要描述两个actor之间如何进行异步通讯，主要定义actor接收到信息之后的行为（Receive），包括处理方式和处理后的消息发送给哪个actor，异步通讯保证了actor的利用率。

```go
type ping struct{ val int }
type pingActor struct{}

func (state *pingActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		fmt.Println("Started, initialize actor here")
	case *actor.Stopping:
		fmt.Println("Stopping, actor is about shut down")
	case *actor.Restarting:
		fmt.Println("Restarting, actor is about restart")
	case *ping:
		val := msg.val
		if val < 10000000 {
			context.Sender().Request(&ping{val: val + 1}, context.Self())
		} else {
			end := time.Now().UnixNano()
			fmt.Printf("%s end %d\n", context.Self().Id, end)
		}
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	props := actor.FromProducer(func() actor.Actor { return &pingActor{} })
	actora := actor.Spawn(props)
	actorb := actor.Spawn(props)
	fmt.Printf("begin time %d\n", time.Now().UnixNano())
	actora.Request(&ping{val: 1}, actorb)
	time.Sleep(10 * time.Second)
	actora.Stop()
	actorb.Stop()
	console.ReadLine()
}
```

## Server/Client 同步调用

本例主要描述如何与actor（server）进行同步通讯，客户端将需求消息发送给actor，并等待actor的返回结果，该需求可能需要多个actor共同协作完成，多个actor之间采用上面例子中的异步通讯来进行处理，最后处理结果返回给client。



#### message.go
```go
type Request struct {
	Who string
}

type Response struct {
	Welcome string
}
```

#### server.go
```go
type Server struct {}

func (server *Server) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		fmt.Println("Started, initialize server actor here")
	case *actor.Stopping:
		fmt.Println("Stopping, actor is about shut down")
	case *actor.Restarting:
		fmt.Println("Restarting, actor is about restart")
	case *message.Request:
		fmt.Println("Receive message", msg.Who)
		context.Sender().Request(&message.Response{Welcome: "Welcome!"}, context.Self())
	}
}

func (server *Server) Start() *actor.PID{
	props := actor.FromProducer(func() actor.Actor { return &Server{} })
	pid := actor.Spawn(props)
	return pid
}

func (server *Server) Stop(pid *actor.PID) {
	pid.Stop()
}
```

#### client.go
```go
type Client struct {}


//Call the server synchronously
func (client *Client) SyncCall(serverPID *actor.PID) (interface{}, error) {
	future := serverPID.RequestFuture(&message.Request{Who: "Ontology"}, 10*time.Second)
	result, err := future.Result()
	return result, err
}
```

#### main.go
```go
func main() {
	server := &server.Server{}
	client := &client.Client{}
	serverPID := server.Start()
	result, err := client.SyncCall(serverPID)
	if err != nil {
		fmt.Println("ERROR:", err)
	}
	fmt.Println(result)
}
```
## EventHub
Actor可以通过EventHub 进行广播和订阅操作，支持ALL,ROUNDROBIN,RANDOM的广播模式

### Example
```go
package main

import (
	"github.com/OntologyNetwork/protoactor-go/eventhub"
	"fmt"
	"github.com/OntologyNetwork/protoactor-go/actor"

	"time"
)


type PubMessage struct{
	message string
}

type ResponseMessage struct{
	message string
}

func main() {

	eh:= eventhub.GlobalEventHub
	subprops := actor.FromFunc(func(context actor.Context) {
		switch msg := context.Message().(type) {

		case PubMessage:
			fmt.Println(context.Self().Id + " get message "+msg.message)
			context.Sender().Request(ResponseMessage{"response message from "+context.Self().Id },context.Self())

		default:

		}
	})

	pubprops := actor.FromFunc(func(context actor.Context) {
		switch msg := context.Message().(type) {

		case ResponseMessage:
			fmt.Println(context.Self().Id + " get message "+msg.message)
			//context.Sender().Request(ResponseMessage{"response message from "+context.Self().Id },context.Self())

		default:
			//fmt.Println("unknown message type")
		}
	})


	publisher, _ := actor.SpawnNamed(pubprops, "publisher")
	sub1, _ := actor.SpawnNamed(subprops, "sub1")
	sub2, _ := actor.SpawnNamed(subprops, "sub2")
	sub3, _ := actor.SpawnNamed(subprops, "sub3")

	topic:= "TEST"
	eh.Subscribe(topic,sub1)
	eh.Subscribe(topic,sub2)
	eh.Subscribe(topic,sub3)

	event := eventhub.Event{Publisher:publisher,Message:PubMessage{"hello fellows"},Topic:topic,Policy:eventhub.PUBLISH_POLICY_ALL}
	eh.Publish(event)
	time.Sleep(2*time.Second)
	fmt.Println("before unsubscribe sleeping...")
	eh.Unsubscribe(topic,sub2)
	eh.Publish(event)
	time.Sleep(2*time.Second)

	fmt.Println("random event...")
	randomevent := eventhub.Event{Publisher:publisher,Message:PubMessage{"hello fellows"},Topic:topic,Policy:eventhub.PUBLISH_POLICY_RANDOM}
	for i:=0 ;i<10;i++{
		eh.Publish(randomevent)
	}

	time.Sleep(2*time.Second)

	fmt.Println("roundrobin event...")
	roundevent := eventhub.Event{Publisher:publisher,Message:PubMessage{"hello fellows"},Topic:topic,Policy:eventhub.PUBLISH_POLICY_ROUNDROBIN}
	for i:=0 ;i<10;i++{
		eh.Publish(roundevent)
	}

	time.Sleep(2*time.Second)
}


```

# 跨节点通讯
本项目采用两种方式实现跨节点通讯，分别是grpc和zeromq，对应于项目中的remote和zmqremote包，使用过程中请根据需求选择所需导入的包（接口是一样的）。
## 性能

## Benchmark
node2/main.go
```go
func main() {
	log.Init()
	runtime.GOMAXPROCS(runtime.NumCPU() * 1)
	runtime.GC()

	remote.Start("127.0.0.1:8080")
	var sender *actor.PID
	props := actor.
		FromFunc(
			func(context actor.Context) {
				switch msg := context.Message().(type) {
				case *messages.StartRemote:
					fmt.Println("Starting")
					sender = msg.Sender
					context.Respond(&messages.Start{})
				case *messages.Ping:
					sender.Tell(&messages.Pong{})
				}
			}).
		WithMailbox(mailbox.Bounded(1000000))
	actor.SpawnNamed(props, "remote")
	for{
		time.Sleep(1 * time.Second)
	}
}
```

node1/main.go
```go
type localActor struct {
	count        int
	wgStop       *sync.WaitGroup
	messageCount int
}

func (state *localActor) Receive(context actor.Context) {
	switch context.Message().(type) {
	case *messages.Pong:
		state.count++
		if state.count%50000 == 0 {
			fmt.Println(state.count)
		}
		if state.count == state.messageCount {
			state.wgStop.Done()
		}
	}
}

func newLocalActor(stop *sync.WaitGroup, messageCount int) actor.Producer {
	return func() actor.Actor {
		return &localActor{
			wgStop:       stop,
			messageCount: messageCount,
		}
	}
}

func main() {
	log.Init()
	runtime.GOMAXPROCS(runtime.NumCPU() * 1)
	runtime.GC()

	var wg sync.WaitGroup

	messageCount := 50000
	//remote.DefaultSerializerID = 1
	remote.Start("127.0.0.1:8081")

	props := actor.
		FromProducer(newLocalActor(&wg, messageCount)).
		WithMailbox(mailbox.Bounded(1000000))

	pid := actor.Spawn(props)

	remotePid := actor.NewPID("127.0.0.1:8080", "remote")
	remotePid.
		RequestFuture(&messages.StartRemote{
			Sender: pid,
		}, 5*time.Second).
		Wait()

	wg.Add(1)

	start := time.Now()
	fmt.Println("Starting to send")

	bb := bytes.NewBuffer([]byte(""))
	for i := 0; i < 2000; i++ {
		bb.WriteString("1234567890")
	}
	message := &messages.Ping{Data: bb.Bytes()}
	for i := 0; i < messageCount; i++ {
		remotePid.Tell(message)
	}

	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("Elapsed %s", elapsed)

	x := int(float32(messageCount*2) / (float32(elapsed) / float32(time.Second)))
	fmt.Printf("Msg per sec %v", x)
}
```

messages/protos.proto

protobuf文件的生成命令：
protoc -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf/ --gogoslick_out=plugins=grpc:. /path/to/protos.proto

```text
syntax = "proto3";
package messages;
import "github.com/Ontology/eventbus/actor/protos.proto";

message Start {}
message StartRemote {
    actor.PID Sender = 1;
}
message Ping {
    bytes Data = 1;
}
message Pong {}
```

## ONT验签测试
见example/testRemoteCrypto目录

This Module is based on AsynkronIT/protoactor-go project, more details goes to https://github.com/AsynkronIT/protoactor-go.