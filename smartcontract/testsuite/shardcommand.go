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
package testsuite

import (
	"github.com/ontio/ontology/common"
	"io"
)

type CmdType = uint32

const NotifyCmd = 0
const InvokeCmd = 1
const GreetCmd = 2
const MultiCmd = 3

type ShardCommand interface {
	common.Serializable
	Deserialization(source *common.ZeroCopySource) error
	CmdType() CmdType
}

func EncodeShardCommandToBytes(cmd ShardCommand) []byte {
	sink := common.NewZeroCopySink(0)
	sink.WriteUint32(cmd.CmdType())
	cmd.Serialization(sink)
	return sink.Bytes()
}

func EncodeShardCommand(sink *common.ZeroCopySink, cmd ShardCommand) {
	sink.WriteUint32(cmd.CmdType())
	cmd.Serialization(sink)
}

func DecodeShardCommand(source *common.ZeroCopySource) (ShardCommand, error) {
	cmdType, eof := source.NextUint32()
	if eof {
		return nil, io.ErrUnexpectedEOF
	}
	switch cmdType {
	case MultiCmd:
		cmd := &MutliCommand{}
		err := cmd.Deserialization(source)
		return cmd, err
	case NotifyCmd:
		cmd := &NotifyCommand{}
		err := cmd.Deserialization(source)
		return cmd, err
	case InvokeCmd:
		cmd := &InvokeCommand{}
		err := cmd.Deserialization(source)
		return cmd, err
	case GreetCmd:
		cmd := &GreetCommand{}
		err := cmd.Deserialization(source)
		return cmd, err
	default:
		panic("unkown cmd")
	}
}

type GreetCommand struct{}

func (self *GreetCommand) CmdType() CmdType {
	return GreetCmd
}

func (self *GreetCommand) Serialization(sink *common.ZeroCopySink) {}
func (self *GreetCommand) Deserialization(source *common.ZeroCopySource) error {
	return nil
}

type NotifyCommand struct {
	Target common.ShardID
	Cmd    ShardCommand
}

func (self *NotifyCommand) CmdType() CmdType {
	return NotifyCmd
}

func (self *NotifyCommand) Serialization(sink *common.ZeroCopySink) {
	sink.WriteShardID(self.Target)
	EncodeShardCommand(sink, self.Cmd)
}

func (self *NotifyCommand) Deserialization(source *common.ZeroCopySource) error {
	id, err := source.NextShardID()
	if err != nil {
		return err
	}
	cmd, err := DecodeShardCommand(source)
	if err != nil {
		return err
	}

	self.Target = id
	self.Cmd = cmd

	return nil
}

type MutliCommand struct {
	SubCmds []ShardCommand
}

func (self MutliCommand) SubCmd(cmd ShardCommand) MutliCommand {
	return MutliCommand{append(self.SubCmds, cmd)}
}

func (self *MutliCommand) CmdType() CmdType {
	return MultiCmd
}

func (self *MutliCommand) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(uint32(len(self.SubCmds)))
	for _, cmd := range self.SubCmds {
		EncodeShardCommand(sink, cmd)
	}
}

func (self *MutliCommand) Deserialization(source *common.ZeroCopySource) error {
	len, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	for i := uint32(0); i < len; i++ {
		cmd, err := DecodeShardCommand(source)
		if err != nil {
			return err
		}
		self.SubCmds = append(self.SubCmds, cmd)
	}

	return nil
}

type InvokeCommand struct {
	NotifyCommand
}

func (self *InvokeCommand) CmdType() CmdType {
	return InvokeCmd
}

func NewNotifyCommand(target common.ShardID, cmd ShardCommand) *NotifyCommand {
	var res NotifyCommand
	res.Target = target
	res.Cmd = cmd
	return &res
}

func NewInvokeCommand(target common.ShardID, cmd ShardCommand) *InvokeCommand {
	var res InvokeCommand
	res.Target = target
	res.Cmd = cmd
	return &res
}
