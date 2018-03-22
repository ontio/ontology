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

package router

import (
	"log"

	"github.com/Ontology/eventbus/actor"
	"github.com/serialx/hashring"
)

type Hasher interface {
	Hash() string
}

type consistentHashGroupRouter struct {
	GroupRouter
}

type consistentHashPoolRouter struct {
	PoolRouter
}

type hashmapContainer struct {
	hashring  *hashring.HashRing
	routeeMap map[string]*actor.PID
}
type consistentHashRouterState struct {
	hmc *hashmapContainer
}

func (state *consistentHashRouterState) SetRoutees(routees *actor.PIDSet) {
	//lookup from node name to PID
	hmc := hashmapContainer{}
	hmc.routeeMap = make(map[string]*actor.PID)
	nodes := make([]string, routees.Len())
	routees.ForEach(func(i int, pid actor.PID) {
		nodeName := pid.Address + "@" + pid.Id
		nodes[i] = nodeName
		hmc.routeeMap[nodeName] = &pid
	})
	//initialize hashring for mapping message keys to node names
	hmc.hashring = hashring.New(nodes)
	state.hmc = &hmc
}

func (state *consistentHashRouterState) GetRoutees() *actor.PIDSet {
	var routees actor.PIDSet
	for _, v := range state.hmc.routeeMap {
		routees.Add(v)
	}
	return &routees
}

func (state *consistentHashRouterState) RouteMessage(message interface{}) {
	_, uwpMsg, _ := actor.UnwrapEnvelope(message)
	switch msg := uwpMsg.(type) {
	case Hasher:
		key := msg.Hash()
		hmc := state.hmc

		node, ok := hmc.hashring.GetNode(key)
		if !ok {
			log.Printf("[ROUTING] Consistent has router failed to derminate routee: %v", key)
			return
		}
		if routee, ok := hmc.routeeMap[node]; ok {
			routee.Tell(message)
		} else {
			log.Println("[ROUTING] Consisten router failed to resolve node", node)
		}
	default:
		log.Println("[ROUTING] Message must implement router.Hasher", msg)
	}
}

func (state *consistentHashRouterState) InvokeRouterManagementMessage(msg ManagementMessage, sender *actor.PID) {

}

func NewConsistentHashPool(size int) *actor.Props {
	return actor.FromSpawnFunc(spawner(&consistentHashPoolRouter{PoolRouter{PoolSize: size}}))
}

func NewConsistentHashGroup(routees ...*actor.PID) *actor.Props {
	return actor.FromSpawnFunc(spawner(&consistentHashGroupRouter{GroupRouter{Routees: actor.NewPIDSet(routees...)}}))
}

func (config *consistentHashPoolRouter) CreateRouterState() Interface {
	return &consistentHashRouterState{}
}

func (config *consistentHashGroupRouter) CreateRouterState() Interface {
	return &consistentHashRouterState{}
}
