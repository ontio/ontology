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

package utils

import (
	"fmt"
	"github.com/ontio/ontology/core/types"
)

//
// build child-Shard Ontology process command line arguments
//
func BuildShardCommandArgs(cmdArgs map[string]string, shardID types.ShardID, parentPort uint64) ([]string, error) {
	args := make([]string, 0)
	shardArgs := make(map[string]string)
	for _, flag := range CmdFlagsForSharding {
		shardArgs[flag.GetName()] = ""
	}

	// prepare Shard-Configs for child-shard ontology process
	shardArgs[ShardIDFlag.GetName()] = fmt.Sprintf("%d", shardID)
	shardArgs[ShardPortFlag.GetName()] = fmt.Sprintf("%d", uint(parentPort+uint64(shardID.Index())))
	shardArgs[ParentShardPortFlag.GetName()] = fmt.Sprintf("%d", parentPort)

	// copy all args to new shard command, except sharding related flags
	for n, v := range cmdArgs {
		// FIXME: disabled consensusPort flag
		if n == ConsensusPortFlag.GetName() {
			continue
		}

		if _, present := shardArgs[n]; !present {
			// non-shard arguments: copy to child-shard
			args = append(args, "--"+n+"="+v)
		}
	}
	for n, shardCfg := range shardArgs {
		if len(shardCfg) > 0 {
			// shard-arguments
			args = append(args, "--"+n+"="+shardCfg)
		}
	}

	return args, nil
}
