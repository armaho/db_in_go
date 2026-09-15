package db_in_go

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
)

var ErrBadSum = errors.New("bad checksum")

type Entry struct {
	key     []byte
	val     []byte
	deleted bool
}

func (ent *Entry) Encode() []byte {
	data := make([]byte, 4+4+4+1+len(ent.key)+len(ent.val))

	binary.LittleEndian.PutUint32(data[4:8], uint32(len(ent.key)))
	binary.LittleEndian.PutUint32(data[8:12], uint32(len(ent.val)))

	if ent.deleted {
		data[12] = 1
	} else {
		data[12] = 0
	}

	copy(data[13:], ent.key)
	copy(data[13+len(ent.key):], ent.val)

	checksum := crc32.ChecksumIEEE(data[4:])
	binary.LittleEndian.PutUint32(data[:4], checksum)

	return data
}

func (ent *Entry) Decode(r io.Reader) error {
	header := make([]byte, 13)
	if _, err := io.ReadFull(r, header); err != nil {
		return err
	}

	ent.deleted = header[12] == 1

	klen := binary.LittleEndian.Uint32(header[4:8])
	var vlen uint32
	if ent.deleted {
		vlen = 0
	} else {
		vlen = binary.LittleEndian.Uint32(header[8:12])
	}

	data := make([]byte, klen+vlen)
	if _, err := io.ReadFull(r, data); err != nil {
		return err
	}

	h := crc32.NewIEEE()
	h.Write(header[4:])
	h.Write(data)
	if h.Sum32() != binary.LittleEndian.Uint32(header[0:4]) {
		return ErrBadSum
	}

	ent.key = data[:klen]
	if ent.deleted {
		ent.val = nil
	} else {
		ent.val = data[klen:]
	}
	return nil
}
