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
	"context"
	"fmt"
	"log/slog"
	"strings"

	olog "github.com/ontio/ontology/common/log"
)

func NewOntLogHandler() slog.Handler {
	return &OntLogHander{}
}

type OntLogHander struct{}

func (self *OntLogHander) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (self *OntLogHander) Handle(ctx context.Context, record slog.Record) error {
	var kvs []string
	record.Attrs(func(attr slog.Attr) bool {
		kvs = append(kvs, fmt.Sprintf("%s=%s", attr.Key, attr.Value.String()))
		return true
	})
	switch record.Level {
	case slog.LevelDebug:
		olog.Debug(record.Message, strings.Join(kvs, ", "))
	case slog.LevelInfo:
		olog.Info(record.Message, strings.Join(kvs, ", "))
	case slog.LevelWarn:
		olog.Warn(record.Message, strings.Join(kvs, ", "))
	case slog.LevelError:
		olog.Error(record.Message, strings.Join(kvs, ", "))
	}

	return nil
}

func (self *OntLogHander) WithAttrs(attrs []slog.Attr) slog.Handler {
	return self
}

func (self *OntLogHander) WithGroup(name string) slog.Handler {
	return self
}
