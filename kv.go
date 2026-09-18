package db_in_go

import (
	"bytes"
	"slices"
)

type KV struct {
	log Log

	keys [][]byte
	vals [][]byte
}

func (kv *KV) Open() error {
	kv.keys = [][]byte{}
	kv.vals = [][]byte{}
	mem := map[string][]byte{}
	err := kv.log.Open()
	if err != nil {
		return err
	}

	for {
		ent := Entry{}
		eof, err := kv.log.Read(&ent)
		if eof {
			break
		}
		if err != nil {
			return err
		}

		if ent.deleted {
			delete(mem, string(ent.key))
		} else {
			mem[string(ent.key)] = ent.val
		}
	}

	for key, val := range mem {
		kv.keys = append(kv.keys, []byte(key))
		kv.vals = append(kv.vals, val)
	}
	slices.SortFunc(kv.keys, bytes.Compare)
	slices.SortFunc(kv.vals, bytes.Compare)

	return nil
}

func (kv *KV) Close() error {
	return kv.log.fp.Close()
}

func (kv *KV) getIdx(key []byte) (idx int, ok bool) {
	idx, ok = slices.BinarySearchFunc(kv.keys, key, bytes.Compare)
	return
}

func (kv *KV) Get(key []byte) (val []byte, ok bool, err error) {
	if idx, ok := kv.getIdx(key); ok {
		return kv.vals[idx], true, nil
	}
	return nil, false, nil
}

type UpdateMode int

const (
	ModeUpsert UpdateMode = 0 // insert or update
	ModeInsert UpdateMode = 1 // insert new
	ModeUpdate UpdateMode = 2 // update existing
)

func (kv *KV) SetEx(key []byte, val []byte, mode UpdateMode) (updated bool, err error) {
	idx, exists := kv.getIdx(key)

	if mode == ModeInsert && exists {
		return false, nil
	}

	if mode == ModeUpdate && !exists {
		return false, nil
	}

	ent := Entry{
		key:     key,
		val:     val,
		deleted: false,
	}
	err = kv.log.Write(&ent)
	if err != nil {
		return
	}

	if exists {
		updated = !bytes.Equal(kv.vals[idx], val)
		kv.vals[idx] = val
	} else {
		updated = true
		kv.keys = slices.Insert(kv.keys, idx, key)
		kv.vals = slices.Insert(kv.vals, idx, val)
	}
	return

}

func (kv *KV) Set(key []byte, val []byte) (updated bool, err error) {
	return kv.SetEx(key, val, ModeUpsert)
}

func (kv *KV) Del(key []byte) (deleted bool, err error) {
	ent := Entry{
		key:     key,
		val:     []byte{},
		deleted: true,
	}
	err = kv.log.Write(&ent)
	if err != nil {
		return
	}

	idx, deleted := kv.getIdx(key)
	if deleted {
		kv.keys = slices.Delete(kv.keys, idx, idx+1)
		kv.vals = slices.Delete(kv.vals, idx, idx+1)
	}

	return
}
