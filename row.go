package db_in_go

import (
	"errors"
	"fmt"
	"slices"
)

type Schema struct {
	Table string
	Cols  []Column
	PKey  []int
}

type Column struct {
	Name string
	Type CellType
}

type Row []Cell

func (schema *Schema) NewRow() Row {
	return make(Row, len(schema.Cols))
}

func (row Row) EncodeKey(schema *Schema) (key []byte) {
	check(len(row) == len(schema.Cols))

	key = append(key, []byte(schema.Table)...)
	key = append(key, 0x00)
	for i, c := range row {
		if !slices.Contains(schema.PKey, i) {
			continue
		}
		check(c.Type == schema.Cols[i].Type)
		key = c.Encode(key)
	}

	return key
}

func (row Row) EncodeVal(schema *Schema) (val []byte) {
	check(len(row) == len(schema.Cols))

	for i, c := range row {
		if slices.Contains(schema.PKey, i) {
			continue
		}
		check(c.Type == schema.Cols[i].Type)
		val = c.Encode(val)
	}

	return val
}

func (row Row) DecodeKey(schema *Schema, key []byte) (err error) {
	if len(key) < len(schema.Table)+1 {
		return errors.New("row does not belong to this schema")
	}
	if string(key[:len(schema.Table)+1]) != schema.Table+"\x00" {
		return fmt.Errorf(
			"row does not belong to this schema (expected %s, got %s)",
			schema.Table, string(key[:len(schema.Table)]))
	}
	key = key[len(schema.Table)+1:]

	for _, keyIdx := range schema.PKey {
		row[keyIdx].Type = schema.Cols[keyIdx].Type
		key, err = row[keyIdx].Decode(key)
		if err != nil {
			return err
		}
	}

	return nil
}

func (row Row) DecodeVal(schema *Schema, val []byte) (err error) {
	for i, c := range schema.Cols {
		if slices.Contains(schema.PKey, i) {
			continue
		}

		row[i].Type = c.Type
		row[i].Decode(val)
	}

	return nil
}
