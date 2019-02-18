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
package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigGeneration(t *testing.T) {
	defaultConfig := NewOntologyConfig()
	assert.Equal(t, defaultConfig, DefConfig)
}

func TestConfigSerialize(t *testing.T) {
	defaultConfig := NewOntologyConfig()
	buf := new(bytes.Buffer)

	if err := defaultConfig.Serialize(buf); err != nil {
		t.Errorf("serialize config err: %s", err)
	}

	cfgBytes := buf.Bytes()
	cfg2 := &OntologyConfig{}
	if err := cfg2.Deserialize(buf); err != nil {
		t.Errorf("deserialize config err: %s", err)
	}

	buf2 := new(bytes.Buffer)
	if err := cfg2.Serialize(buf2); err != nil {
		t.Errorf("serialize config2 err: %s", err)
	}

	if bytes.Compare(cfgBytes, buf2.Bytes()) != 0 {
		t.Errorf("config ser/des err")
	}
}
