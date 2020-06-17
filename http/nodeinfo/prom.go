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

package nodeinfo

import (
	"time"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/p2pserver/net/netserver"
	p2p "github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/protocols"
	prom "github.com/prometheus/client_golang/prometheus"
)

var (
	nodePortMetric = prom.NewGauge(prom.GaugeOpts{
		Name: "ontology_nodeport",
		Help: "ontology node port",
	})

	blockHeightMetric = prom.NewGauge(prom.GaugeOpts{
		Name: "ontology_block_height",
		Help: "ontology blockchain block height",
	})

	inboundsCountMetric = prom.NewGauge(prom.GaugeOpts{
		Name: "ontology_p2p_inbounds_count",
		Help: "ontology p2p inbloud count",
	})

	outboundsCountMetric = prom.NewGauge(prom.GaugeOpts{
		Name: "ontology_p2p_outbounds_count",
		Help: "ontology p2p outbloud count",
	})

	peerStatusMetric = prom.NewGaugeVec(prom.GaugeOpts{
		Name: "ontology_p2p_peer_status",
		Help: "ontology peer info",
	}, []string{"ip", "id"})

	reconnectCountMetric = prom.NewGauge(prom.GaugeOpts{
		Name: "ontology_p2p_reconnect_count",
		Help: "ontology p2p reconnect count",
	})
)

var (
	metrics = []prom.Collector{nodePortMetric, blockHeightMetric, inboundsCountMetric,
		outboundsCountMetric, peerStatusMetric, reconnectCountMetric}
)

func initMetric() error {
	for _, curMetric := range metrics {
		if err := prom.Register(curMetric); err != nil {
			return err
		}
	}

	return nil
}

func metricUpdate(n p2p.P2P) {
	nodePortMetric.Set(float64(config.DefConfig.P2PNode.NodePort))

	blockHeightMetric.Set(float64(ledger.DefLedger.GetCurrentBlockHeight()))

	ns, ok := n.(*netserver.NetServer)
	if !ok {
		return
	}

	inboundsCountMetric.Set(float64(ns.ConnectController().InboundsCount()))
	outboundsCountMetric.Set(float64(ns.ConnectController().OutboundsCount()))

	peers := ns.GetNeighbors()
	for _, curPeer := range peers {
		id := curPeer.GetID()

		// label: IP PeedID
		peerStatusMetric.WithLabelValues(curPeer.GetAddr(), id.ToHexString()).Set(float64(curPeer.GetHeight()))
	}

	pt := ns.Protocol()
	mh, ok := pt.(*protocols.MsgHandler)
	if !ok {
		return
	}

	reconnectCountMetric.Set(float64(mh.ReconnectService().ReconnectCount()))
}

func updateMetric(n p2p.P2P) {
	tk := time.NewTicker(time.Minute)
	defer tk.Stop()
	for {
		select {
		case <-tk.C:
			metricUpdate(n)
		}
	}
}
