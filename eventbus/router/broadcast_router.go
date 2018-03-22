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

import "github.com/Ontology/eventbus/actor"

type broadcastGroupRouter struct {
	GroupRouter
}

type broadcastPoolRouter struct {
	PoolRouter
}

type broadcastRouterState struct {
	routees *actor.PIDSet
}

func (state *broadcastRouterState) SetRoutees(routees *actor.PIDSet) {
	state.routees = routees
}

func (state *broadcastRouterState) GetRoutees() *actor.PIDSet {
	return state.routees
}

func (state *broadcastRouterState) RouteMessage(message interface{}) {
	state.routees.ForEach(func(i int, pid actor.PID) {
		pid.Tell(message)
	})
}

func NewBroadcastPool(size int) *actor.Props {
	return actor.FromSpawnFunc(spawner(&broadcastPoolRouter{PoolRouter{PoolSize: size}}))
}

func NewBroadcastGroup(routees ...*actor.PID) *actor.Props {
	return actor.FromSpawnFunc(spawner(&broadcastGroupRouter{GroupRouter{Routees: actor.NewPIDSet(routees...)}}))
}

func (config *broadcastPoolRouter) CreateRouterState() Interface {
	return &broadcastRouterState{}
}

func (config *broadcastGroupRouter) CreateRouterState() Interface {
	return &broadcastRouterState{}
}
