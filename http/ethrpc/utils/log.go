package utils

import (
	"github.com/ethereum/go-ethereum/log"
	olog "github.com/ontio/ontology/common/log"
)

func OntLogHandler() log.Handler {
	h := log.FuncHandler(func(r *log.Record) error {
		olog.Error(r.Msg)
		return nil
	})
	return log.LazyHandler(log.SyncHandler(h))
}
