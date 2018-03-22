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
	"sync/atomic"

	"github.com/Ontology/eventbus/actor"
)

type roundRobinGroupRouter struct {
	GroupRouter
}

type roundRobinPoolRouter struct {
	PoolRouter
}

type roundRobinState struct {
	index   int32
	routees *actor.PIDSet
	values  []actor.PID
}

func (state *roundRobinState) SetRoutees(routees *actor.PIDSet) {
	state.routees = routees
	state.values = routees.Values()
}

func (state *roundRobinState) GetRoutees() *actor.PIDSet {
	return state.routees
}

func (state *roundRobinState) RouteMessage(message interface{}) {
	pid := roundRobinRoutee(&state.index, state.values)
	pid.Tell(message)
}

func NewRoundRobinPool(size int) *actor.Props {
	return actor.FromSpawnFunc(spawner(&roundRobinPoolRouter{PoolRouter{PoolSize: size}}))
}

func NewRoundRobinGroup(routees ...*actor.PID) *actor.Props {
	return actor.FromSpawnFunc(spawner(&roundRobinGroupRouter{GroupRouter{Routees: actor.NewPIDSet(routees...)}}))
}

func (config *roundRobinPoolRouter) CreateRouterState() Interface {
	return &roundRobinState{}
}

func (config *roundRobinGroupRouter) CreateRouterState() Interface {
	return &roundRobinState{}
}

func roundRobinRoutee(index *int32, routees []actor.PID) actor.PID {
	i := int(atomic.AddInt32(index, 1))
	if i < 0 {
		*index = 0
		i = 0
	}
	mod := len(routees)
	routee := routees[i%mod]
	return routee
}
