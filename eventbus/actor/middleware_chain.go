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

func makeInboundMiddlewareChain(middleware []InboundMiddleware, lastReceiver ActorFunc) ActorFunc {
	if len(middleware) == 0 {
		return nil
	}

	h := middleware[len(middleware)-1](lastReceiver)
	for i := len(middleware) - 2; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

func makeOutboundMiddlewareChain(outboundMiddleware []OutboundMiddleware, lastSender SenderFunc) SenderFunc {
	if len(outboundMiddleware) == 0 {
		return nil
	}

	h := outboundMiddleware[len(outboundMiddleware)-1](lastSender)
	for i := len(outboundMiddleware) - 2; i >= 0; i-- {
		h = outboundMiddleware[i](h)
	}
	return h
}
