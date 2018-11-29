package creator

import (
	"github.com/ontio/ontology/common/log"
	"testing"
)

func init() {
	log.Init(log.Stdout)
}

func TestFactory(t *testing.T) {
	tcpTransport1, _ := GetTransportFactory().GetTransport("ert")
	if tcpTransport1 != nil {
		t.Error("tcpTransport1 should be nil")
	}

	tcpTransport2, _ := GetTransportFactory().GetTransport("TCP")
	if tcpTransport2 == nil {
		t.Error("tcpTransport2 shouldnot be nil")
	}
}
