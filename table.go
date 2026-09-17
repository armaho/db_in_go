package db_in_go

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
)

type DB struct {
	KV     KV
	tables map[string]Schema
}

type SQLResult struct {
	Updated int
	Header  []string
	Values  []Row
}

func schemaKey(table string) []byte {
	return []byte("@schema_" + table)
}

func (db *DB) GetSchema(table string) (Schema, error) {
	schema, ok := db.tables[table]
	if ok {
		return schema, nil
	}

	val, ok, err := db.KV.Get(schemaKey(table))
	if err != nil {
		return Schema{}, err
	}
	if !ok {
		return Schema{}, errors.New("table not found")
	}

	err = json.Unmarshal(val, &schema)
	if err != nil {
		return Schema{}, err
	}

	db.tables[table] = schema
	return schema, nil
}

func (db *DB) ExecStmt(stmt any) (r SQLResult, err error) {
	switch ptr := stmt.(type) {
	case *StmtCreatTable:
		err = db.execCreateTable(ptr)
	case *StmtSelect:
		r.Header = ptr.cols
		r.Values, err = db.execSelect(ptr)
	case *StmtInsert:
		r.Updated, err = db.execInsert(ptr)
	case *StmtUpdate:
		r.Updated, err = db.execUpdate(ptr)
	case *StmtDelete:
		r.Updated, err = db.execDelete(ptr)
	default:
		panic("unreachable")
	}
	return
}

func (db *DB) execDelete(stmt *StmtDelete) (count int, err error) {
	schema, err := db.GetSchema(stmt.table)
	if err != nil {
		return 0, err
	}

	row, err := makePKey(&schema, stmt.keys)
	if err != nil {
		return 0, err
	}

	if updated, err := db.Delete(&schema, row); err != nil || !updated {
		return 0, err
	}
	return 1, nil
}

func (db *DB) execUpdate(stmt *StmtUpdate) (count int, err error) {
	schema, err := db.GetSchema(stmt.table)
	if err != nil {
		return 0, err
	}

	row, err := makePKey(&schema, stmt.keys)
	if err != nil {
		return 0, err
	}

	// makePKey leaves other columns with default value of Cell.
	// We use that to fill values here.
	for i, cell := range row {
		// This is already filled as it's a primary key
		if cell.Type != 0 {
			continue
		}

		found := false
		for _, ncell := range stmt.value {
			if schema.Cols[i].Name != ncell.column {
				continue
			}

			found = true
			row[i] = ncell.value
			break
		}
		if !found {
			return 0, fmt.Errorf("Cannot find column %s", schema.Cols[i].Name)
		}
	}

	if updated, err := db.Update(&schema, row); err != nil || !updated {
		return 0, err
	}
	return 1, nil
}

func (db *DB) execInsert(stmt *StmtInsert) (count int, err error) {
	schema, err := db.GetSchema(stmt.table)
	if err != nil {
		return 0, err
	}

	updated, err := db.Insert(&schema, stmt.value)
	if err != nil || !updated {
		return 0, err
	}
	return 1, nil
}

func (db *DB) execCreateTable(stmt *StmtCreatTable) (err error) {
	schema := Schema{
		Table: stmt.table,
		Cols:  []Column{},
		PKey:  []int{},
	}

	for _, c := range stmt.cols {
		schema.Cols = append(schema.Cols, Column{
			Type: c.Type,
			Name: c.Name,
		})
	}

	schema.PKey, err = lookupColumns(schema.Cols, stmt.pkey)
	if err != nil {
		return err
	}

	val, err := json.Marshal(schema)
	if err != nil {
		return err
	}

	_, err = db.KV.Set(schemaKey(schema.Table), val)
	if err != nil {
		return err
	}

	return nil
}

func lookupColumns(cols []Column, selected []string) (
	selectedIdx []int, err error) {
	for _, colName := range selected {
		found := false
		for i, col := range cols {
			if col.Name == colName {
				selectedIdx = append(selectedIdx, i)
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("Invalid column name %s", colName)
		}
	}
	return
}

func makePKey(schema *Schema, keys []NamedCell) (row Row, err error) {
	for i, col := range schema.Cols {
		if !slices.Contains(schema.PKey, i) {
			row = append(row, Cell{})
			continue
		}

		found := false
		for _, namedCell := range keys {
			if namedCell.column != col.Name {
				continue
			}

			row = append(row, namedCell.value)
			found = true
			break
		}

		if !found {
			return nil, fmt.Errorf("Cannot find primary key %s", col.Name)
		}
	}

	return
}

func subsetRow(row Row, selected []int) (res Row) {
	for _, i := range selected {
		res = append(res, row[i])
	}
	return
}

func (db *DB) execSelect(stmt *StmtSelect) ([]Row, error) {
	schema, err := db.GetSchema(stmt.table)
	if err != nil {
		return nil, err
	}

	indices, err := lookupColumns(schema.Cols, stmt.cols)
	if err != nil {
		return nil, err
	}
	row, err := makePKey(&schema, stmt.keys)
	if err != nil {
		return nil, err
	}
	if ok, err := db.Select(&schema, row); err != nil || !ok {
		return nil, err
	}

	row = subsetRow(row, indices)
	return []Row{row}, nil
}

func (db *DB) Open() error {
	db.tables = map[string]Schema{}
	return db.KV.Open()
}

func (db *DB) Close() error { return db.KV.Close() }

func (db *DB) Select(schema *Schema, row Row) (ok bool, err error) {
	key := row.EncodeKey(schema)
	val, ok, err := db.KV.Get(key)
	if err != nil || !ok {
		return
	}

	err = row.DecodeVal(schema, val)
	if err != nil {
		return
	}
	return true, nil
}

func (db *DB) Insert(schema *Schema, row Row) (updated bool, err error) {
	key := row.EncodeKey(schema)
	val := row.EncodeVal(schema)

	return db.KV.SetEx(key, val, ModeInsert)
}

func (db *DB) Upsert(schema *Schema, row Row) (updated bool, err error) {
	key := row.EncodeKey(schema)
	val := row.EncodeVal(schema)

	return db.KV.Set(key, val)
}

func (db *DB) Update(schema *Schema, row Row) (updated bool, err error) {
	key := row.EncodeKey(schema)
	val := row.EncodeVal(schema)

	return db.KV.SetEx(key, val, ModeUpdate)
}

func (db *DB) Delete(schema *Schema, row Row) (deleted bool, err error) {
	key := row.EncodeKey(schema)
	return db.KV.Del(key)
}
