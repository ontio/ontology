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
	"bytes"
	"compress/zlib"
	"fmt"
	"github.com/ontio/ontology/common/serialization"
	"io"
	"io/ioutil"
)

const (
	COMPRESS_TYPE_ZLIB = iota
)

const (
	DEFAULT_COMPRESS_TYPE         = COMPRESS_TYPE_ZLIB
	EXPORT_BLOCK_METADATA_LEN     = 256
	EXPORT_BLOCK_METADATA_VERSION = 1
)

type ExportBlockMetadata struct {
	Version      byte
	CompressType byte
	BlockHeight  uint32
}

func NewExportBlockMetadata() *ExportBlockMetadata {
	return &ExportBlockMetadata{
		Version:      EXPORT_BLOCK_METADATA_VERSION,
		CompressType: DEFAULT_COMPRESS_TYPE,
	}
}

func (this *ExportBlockMetadata) Serialize(w io.Writer) error {
	metadata := make([]byte, EXPORT_BLOCK_METADATA_LEN, EXPORT_BLOCK_METADATA_LEN)
	buf := bytes.NewBuffer(nil)
	err := serialization.WriteByte(buf, this.Version)
	if err != nil {
		return err
	}
	err = serialization.WriteByte(buf, this.CompressType)
	if err != nil {
		return err
	}
	err = serialization.WriteUint32(buf, this.BlockHeight)
	data := buf.Bytes()
	if len(data) > EXPORT_BLOCK_METADATA_LEN {
		return fmt.Errorf("metata len size larger than %d", EXPORT_BLOCK_METADATA_LEN)
	}
	copy(metadata, data)
	_, err = w.Write(metadata)
	return err
}

func (this *ExportBlockMetadata) Deserialize(r io.Reader) error {
	metadata := make([]byte, EXPORT_BLOCK_METADATA_LEN, EXPORT_BLOCK_METADATA_LEN)
	_, err := io.ReadFull(r, metadata)
	if err != nil {
		return err
	}
	if metadata[0] != EXPORT_BLOCK_METADATA_VERSION {
		return fmt.Errorf("version unmatch")
	}
	reader := bytes.NewBuffer(metadata)
	ver, err := serialization.ReadByte(reader)
	if err != nil {
		return err
	}
	this.Version = ver
	compressType, err := serialization.ReadByte(reader)
	if err != nil {
		return err
	}
	this.CompressType = compressType
	height, err := serialization.ReadUint32(reader)
	if err != nil {
		return err
	}
	this.BlockHeight = height
	return nil
}

func CompressBlockData(data []byte, compressType byte) ([]byte, error) {
	switch compressType {
	case COMPRESS_TYPE_ZLIB:
		return ZLibCompress(data)
	default:
		return nil, fmt.Errorf("unknown compress type")
	}
}

func DecompressBlockData(data []byte, compressType byte) ([]byte, error) {
	switch compressType {
	case COMPRESS_TYPE_ZLIB:
		return ZLibDecompress(data)
	default:
		return nil, fmt.Errorf("unknown compress type")
	}
}

func ZLibCompress(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	zlibWriter := zlib.NewWriter(buf)
	_, err := zlibWriter.Write(data)
	if err != nil {
		return nil, fmt.Errorf("zlibWriter.Write error %s", err)
	}
	zlibWriter.Close()
	return buf.Bytes(), nil
}

func ZLibDecompress(data []byte) ([]byte, error) {
	buf := bytes.NewReader(data)
	zlibReader, err := zlib.NewReader(buf)
	if err != nil {
		return nil, fmt.Errorf("zlib.NewReader error %s", err)
	}
	defer zlibReader.Close()

	return ioutil.ReadAll(zlibReader)
}
