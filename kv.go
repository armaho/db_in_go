package db_in_go

import "bytes"

type KV struct {
	log Log
	mem map[string][]byte
}

func (kv *KV) Open() error {
	kv.mem = map[string][]byte{}
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
			print(err.Error())
			return err
		}

		if ent.deleted {
			delete(kv.mem, string(ent.key))
		} else {
			kv.mem[string(ent.key)] = ent.val
		}
	}

	return nil
}

func (kv *KV) Close() error {
	return kv.log.fp.Close()
}

func (kv *KV) Get(key []byte) (val []byte, ok bool, err error) {
	val, ok = kv.mem[string(key)]
	return
}

func (kv *KV) Set(key []byte, val []byte) (updated bool, err error) {
	ent := Entry{
		key:     key,
		val:     val,
		deleted: false,
	}
	err = kv.log.Write(&ent)
	if err != nil {
		return
	}

	prev, exists := kv.mem[string(key)]
	kv.mem[string(key)] = val
	updated = !exists || bytes.Equal(prev, val)
	return
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

	_, deleted = kv.mem[string(key)]
	delete(kv.mem, string(key))
	return
}
