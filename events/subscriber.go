package events

type EventFunc func(v interface{})

type Subscriber chan interface{}
