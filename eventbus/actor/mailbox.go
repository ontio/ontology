package actor

import "github.com/Ontology/eventbus/mailbox"

var (
	defaultDispatcher = mailbox.NewDefaultDispatcher(300)
)

var defaultMailboxProducer = mailbox.Unbounded()
