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

package log

import (
	"testing"
)

func TestDebugPrint(t *testing.T) {
	CreatePrintLog("./")
	Debug("debug testing")
}

func TestInfoPrint(t *testing.T) {
	CreatePrintLog("./")
	Info("Info testing")
}

func TestWarningPrint(t *testing.T) {
	CreatePrintLog("./")
	Warn("Warning testing")
}

func TestErrorPrint(t *testing.T) {
	CreatePrintLog("./")
	Error("Error testing")
}

func TestFatalPrint(t *testing.T) {
	CreatePrintLog("./")
	Fatal("Fatal testing")
}
