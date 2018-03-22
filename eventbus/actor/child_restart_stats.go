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

package actor

import (
	"time"
)

//RestartStatistics keeps track of how many times an actor have restarted and when
type RestartStatistics struct {
	failureTimes []time.Time
}

//NewRestartStatistics construct a RestartStatistics
func NewRestartStatistics() *RestartStatistics {
	return &RestartStatistics{[]time.Time{}}
}

//FailureCount returns failure count
func (rs *RestartStatistics) FailureCount() int {
	return len(rs.failureTimes)
}

//Fail increases the associated actors failure count
func (rs *RestartStatistics) Fail() {
	rs.failureTimes = append(rs.failureTimes, time.Now())
}

//Reset the associated actors failure count
func (rs *RestartStatistics) Reset() {
	rs.failureTimes = []time.Time{}
}

//NumberOfFailures returns number of failures within a given duration
func (rs *RestartStatistics) NumberOfFailures(withinDuration time.Duration) int {
	if withinDuration == 0 {
		return len(rs.failureTimes)
	}

	num := 0
	currTime := time.Now()
	for _, t := range rs.failureTimes {
		if currTime.Sub(t) < withinDuration {
			num++
		}
	}
	return num
}
