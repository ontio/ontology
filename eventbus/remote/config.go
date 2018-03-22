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

package remote

import "google.golang.org/grpc"

//RemotingOption configures how the remote infrastructure is started
type RemotingOption func(*remoteConfig)

func defaultRemoteConfig() *remoteConfig {
	return &remoteConfig{
		dialOptions:              []grpc.DialOption{grpc.WithInsecure()},
		endpointWriterBatchSize:  1,
		endpointManagerBatchSize: 1,
		endpointWriterQueueSize:  1000000,
		endpointManagerQueueSize: 1000000,
	}
}

func WithEndpointWriterBatchSize(batchSize int) RemotingOption {
	return func(config *remoteConfig) {
		config.endpointWriterBatchSize = batchSize
	}
}

func WithEndpointWriterQueueSize(queueSize int) RemotingOption {
	return func(config *remoteConfig) {
		config.endpointWriterQueueSize = queueSize
	}
}

func WithEndpointManagerBatchSize(batchSize int) RemotingOption {
	return func(config *remoteConfig) {
		config.endpointManagerBatchSize = batchSize
	}
}

func WithEndpointManagerQueueSize(queueSize int) RemotingOption {
	return func(config *remoteConfig) {
		config.endpointManagerQueueSize = queueSize
	}
}

func WithDialOptions(options ...grpc.DialOption) RemotingOption {
	return func(config *remoteConfig) {
		config.dialOptions = options
	}
}

func WithServerOptions(options ...grpc.ServerOption) RemotingOption {
	return func(config *remoteConfig) {
		config.serverOptions = options
	}
}

func WithCallOptions(options ...grpc.CallOption) RemotingOption {
	return func(config *remoteConfig) {
		config.callOptions = options
	}
}

type remoteConfig struct {
	serverOptions            []grpc.ServerOption
	callOptions              []grpc.CallOption
	dialOptions              []grpc.DialOption
	endpointWriterBatchSize  int
	endpointWriterQueueSize  int
	endpointManagerBatchSize int
	endpointManagerQueueSize int
}
