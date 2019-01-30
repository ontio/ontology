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
)

func BuildShardCommandArgs(cmdArgs map[string]string, shardID, parentShardID, parentPort uint64) ([]string, error) {
	args := make([]string, 0)
	shardArgs := make(map[string]string)
	for _, flag := range CmdFlagsForSharding {
		shardArgs[flag.GetName()] = ""
	}
	shardArgs[ShardIDFlag.GetName()] = fmt.Sprintf("%d", shardID)
	shardArgs[ShardPortFlag.GetName()] = fmt.Sprintf("%d",  uint(parentPort + shardID - parentShardID))
	shardArgs[ParentShardIDFlag.GetName()] = fmt.Sprintf("%d", parentShardID)
	shardArgs[ParentShardPortFlag.GetName()] = fmt.Sprintf("%d", parentPort)

	// copy all args to new shard command, except sharding related flags
	for n, v := range cmdArgs {
		if n == ConsensusPortFlag.GetName() {
			continue
		}

		if shardCfg, present := shardArgs[n]; !present {
			args = append(args, "--" + n+"="+v)
		} else if len(shardCfg) > 0 {
			args = append(args, "--" + n+"="+shardCfg)
		}
	}

	return args, nil
}
