package events


type EventType int16

const (
	EventSaveBlock                  EventType = 0
	EventReplyTx                    EventType = 1
)
