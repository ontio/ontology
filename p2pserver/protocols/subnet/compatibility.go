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

package subnet

import (
	"github.com/blang/semver"
)

const MIN_VERSION_FOR_SUBNET = "2.0.0-0"

func supportSubnet(version string) bool {
	if version == "" {
		return false
	}
	v1, err := semver.ParseTolerant(version)
	if err != nil {
		return false
	}
	min, err := semver.ParseTolerant(MIN_VERSION_FOR_SUBNET)
	if err != nil {
		panic(err) // enforced by testcase
	}

	return v1.GTE(min)
}
