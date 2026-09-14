/*
 * Copyright (C) 2024 The ontology Authors
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

package httpserver

import (
	"net/http"
	"testing"
)

// TestNewEnablesReadAndIdleTimeouts guards against the protection being
// silently disabled: net/http treats a zero timeout as "no limit", so a
// regression to zero would reintroduce the slow-client resource exhaustion.
func TestNewEnablesReadAndIdleTimeouts(t *testing.T) {
	srv := New(http.NewServeMux())
	if srv.ReadHeaderTimeout <= 0 {
		t.Error("ReadHeaderTimeout must be set")
	}
	if srv.ReadTimeout <= 0 {
		t.Error("ReadTimeout must be set")
	}
	if srv.IdleTimeout <= 0 {
		t.Error("IdleTimeout must be set")
	}
	if ReadHeaderTimeout > ReadTimeout {
		t.Errorf("ReadHeaderTimeout (%v) should not exceed ReadTimeout (%v)", ReadHeaderTimeout, ReadTimeout)
	}
	if srv.WriteTimeout != 0 {
		t.Error("WriteTimeout should stay unset so long running handlers are not cut off")
	}
}
