package types

import (
	"bytes"
	"github.com/ontio/ontology/common/log"
	"sync"
	"time"
)

type DHTRequestType string

const (
	DHT_FIND_NODE_REQUEST DHTRequestType = "find_node"
	DHT_PING_REQUEST      DHTRequestType = "ping"
)

const MESSAGE_POOL_BUFFER_SIZE = 4

type RequestId string

func ConstructRequestId(nodeId NodeID, reqType DHTRequestType) RequestId {
	var buffer bytes.Buffer
	buffer.WriteString(nodeId.String())
	buffer.WriteString(string(reqType))
	return RequestId(buffer.String())
}

func GetReqTypeFromReqId(reqId RequestId) DHTRequestType {
	nodeIdLength := len(NodeID{}.String())
	return DHTRequestType(reqId[nodeIdLength:])
}

type DHTMessagePool struct {
	lock sync.RWMutex

	requestTimerQueue map[RequestId]*time.Timer
	timeoutListener   chan RequestId
	onTimeOut         func(id RequestId) // time out event should be handled by dht

	resultChan  chan []*Node
	requestPool map[RequestId]*Node
	replaceNode map[RequestId]*Node

	waitGroup map[RequestId]*sync.WaitGroup // used to synchronize
}

func NewRequestPool(onTimeOut func(id RequestId)) *DHTMessagePool {
	msgPool := new(DHTMessagePool)
	msgPool.requestTimerQueue = make(map[RequestId]*time.Timer)
	msgPool.timeoutListener = make(chan RequestId, MESSAGE_POOL_BUFFER_SIZE)
	msgPool.onTimeOut = onTimeOut
	msgPool.resultChan = make(chan []*Node, MESSAGE_POOL_BUFFER_SIZE)
	msgPool.requestPool = make(map[RequestId]*Node)
	msgPool.replaceNode = make(map[RequestId]*Node)
	msgPool.waitGroup = make(map[RequestId]*sync.WaitGroup)
	go msgPool.start()
	return msgPool
}

func (this *DHTMessagePool) start() {
	for {
		select {
		// time out
		case requestId := <-this.timeoutListener:
			this.onTimeOut(requestId)
		}
	}
}

// AddRequest: when send a ping or find node request, call this function
// destinateNode: request to the node
// reqType: request type
// supportData: store some data to support this request
// shouldWait: the request should be waited or not, if is true, must call Wait func
func (this *DHTMessagePool) AddRequest(destinateNode *Node, reqType DHTRequestType, replaceNode *Node,
	waitGroup *sync.WaitGroup) (id RequestId, isNewRequest bool) {
	this.lock.Lock()
	defer this.lock.Unlock()

	requestId := ConstructRequestId(destinateNode.ID, reqType)
	var timeout time.Duration
	if reqType == DHT_FIND_NODE_REQUEST {
		timeout = FIND_NODE_TIMEOUT
	} else if reqType == DHT_PING_REQUEST {
		timeout = PING_TIMEOUT
	} else {
		timeout = DEFAULT_TIMEOUT
	}
	_, ok := this.requestPool[requestId]
	if ok { // if request already exist, reset timer
		log.Info("reset old request:", requestId)
		this.requestTimerQueue[requestId].Reset(timeout)
	} else { // add a new request to pool
		log.Info("send new request:", requestId)
		this.requestPool[requestId] = destinateNode
		this.replaceNode[requestId] = replaceNode
		if waitGroup != nil {
			this.waitGroup[requestId] = waitGroup
			waitGroup.Add(1)
		}

		timer := time.AfterFunc(timeout, func() {
			this.timeoutListener <- requestId
		})
		this.requestTimerQueue[requestId] = timer
		go func() {
			<-timer.C
		}()
	}
	return requestId, !ok
}

func (this *DHTMessagePool) GetReplaceNode(id RequestId) (*Node, bool) {
	this.lock.RLock()
	defer this.lock.RUnlock()

	node, ok := this.replaceNode[id]
	return node, ok
}

func (this *DHTMessagePool) GetRequestData(id RequestId) (*Node, bool) {
	this.lock.RLock()
	defer this.lock.RUnlock()

	node, ok := this.requestPool[id]
	return node, ok
}

func (this *DHTMessagePool) DeleteRequest(requestId RequestId) {
	this.lock.Lock()
	defer this.lock.Unlock()

	_, ok := this.requestPool[requestId]
	if ok {
		delete(this.requestPool, requestId)
		delete(this.replaceNode, requestId)
		// is synchronized request
		if _, ok := this.waitGroup[requestId]; ok {
			this.waitGroup[requestId].Done()
			delete(this.waitGroup, requestId)
		}
	} else {
		return
	}
	timer, ok := this.requestTimerQueue[requestId]
	if ok {
		delete(this.requestTimerQueue, requestId)
	}
	if timer != nil {
		timer.Stop()
	}
}

// push result
func (this *DHTMessagePool) SetResults(results []*Node) {
	this.resultChan <- results
}

// get results channel
func (this *DHTMessagePool) GetResultChan() <-chan []*Node {
	return this.resultChan
}
