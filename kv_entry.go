package db_in_go

import (
	"encoding/binary"
	"io"
)

type Entry struct {
	key     []byte
	val     []byte
	deleted bool
}

func (ent *Entry) Encode() []byte {
	data := make([]byte, 4+4+1+len(ent.key)+len(ent.val))

	binary.LittleEndian.PutUint32(data[0:4], uint32(len(ent.key)))
	binary.LittleEndian.PutUint32(data[4:8], uint32(len(ent.val)))

	if ent.deleted {
		data[8] = 1
	} else {
		data[8] = 0
	}

	copy(data[9:], ent.key)
	copy(data[9+len(ent.key):], ent.val)

	return data
}

func (ent *Entry) Decode(r io.Reader) error {
	header := make([]byte, 9)
	if _, err := io.ReadFull(r, header); err != nil {
		return err
	}

	ent.deleted = header[8] == 1

	klen := binary.LittleEndian.Uint32(header[0:4])
	var vlen uint32
	if ent.deleted {
		vlen = 0
	} else {
		vlen = binary.LittleEndian.Uint32(header[4:8])
	}

	data := make([]byte, klen+vlen)
	if _, err := io.ReadFull(r, data); err != nil {
		return err
	}

	ent.key = data[:klen]
	if ent.deleted {
		ent.val = nil
	} else {
		ent.val = data[klen:]
	}
	return nil
}
