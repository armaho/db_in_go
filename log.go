package db_in_go

import (
	"io"
	"os"
	"path"
	"syscall"
)

type Log struct {
	FileName string
	fp       *os.File
}

func createFileSync(file string) (*os.File, error) {
	fp, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}
	if err = syncDir(path.Base(file)); err != nil {
		_ = fp.Close()
		return nil, err
	}
	return fp, nil
}

func syncDir(file string) error {
	flags := os.O_RDONLY | syscall.O_DIRECTORY
	dirfd, err := syscall.Open(path.Dir(file), flags, 0o644)
	if err != nil {
		return err
	}
	defer syscall.Close(dirfd)
	return syscall.Fsync(dirfd)
}

func (l *Log) Open() (err error) {
	l.fp, err = createFileSync(l.FileName)
	return err
}

func (l *Log) Close() error {
	return l.fp.Close()
}

func (l *Log) Read(ent *Entry) (eof bool, err error) {
	err = ent.Decode(l.fp)
	if err == io.EOF || err == ErrBadSum || err == io.ErrUnexpectedEOF {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

func (l *Log) Write(ent *Entry) error {
	if _, err := l.fp.Write(ent.Encode()); err != nil {
		return err
	}
	return l.fp.Sync()
}
