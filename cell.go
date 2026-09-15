package db_in_go

import (
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
)

type CellType uint8

const (
	TypeI64 CellType = 1
	TypeStr CellType = 2
)

type Cell struct {
	Type CellType
	I64  int64
	Str  []byte
}

func (c *Cell) Encode(toAppend []byte) []byte {
	switch c.Type {
	case TypeI64:
		return binary.LittleEndian.AppendUint64(toAppend, uint64(c.I64))
	case TypeStr:
		toAppend = binary.LittleEndian.AppendUint32(toAppend, uint32(len(c.Str)))
		return append(toAppend, c.Str...)
	default:
		panic("invalid type for cell")
	}
}

func (c *Cell) Decode(data []byte) (rest []byte, err error) {
	switch c.Type {
	case TypeI64:
		if len(data) < 8 {
			return data, errors.New("expected at least 8 bytes for I64 cell")
		}

		c.I64 = int64(binary.LittleEndian.Uint64(data[:8]))
		return data[8:], nil
	case TypeStr:
		if len(data) < 4 {
			return data, errors.New("expected at least 4 bytes for Str cell")
		}

		length := int(binary.LittleEndian.Uint32(data[:4]))
		if len(data) < 4+length {
			return data, fmt.Errorf("expected an str of len %d, got %d",
				length, len(data)-4)
		}
		c.Str = slices.Clone(data[4 : length+4])

		return data[length+4:], nil
	default:
		panic("invalid type for cell")
	}
}
