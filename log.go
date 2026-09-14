package db_in_go

import (
	"io"
	"os"
)

type Log struct {
	FileName string
	fp       *os.File
}

func (l *Log) Open() (err error) {
	l.fp, err = os.OpenFile(l.FileName, os.O_CREATE|os.O_RDWR, 0o644)
	return err
}

func (l *Log) Close() error {
	return l.fp.Close()
}

func (l *Log) Read(ent *Entry) (eof bool, err error) {
	err = ent.Decode(l.fp)
	if err == io.EOF {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

func (l *Log) Write(ent *Entry) error {
	_, err := l.fp.Write(ent.Encode())
	return err
}
