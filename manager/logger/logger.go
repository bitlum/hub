package logger

import (
	"github.com/golang/protobuf/proto"
	"github.com/go-errors/errors"
	"io"
	"encoding/binary"
)

func WriteLogs(w io.Writer, logs []*Log) error {
	for _, log := range logs {
		if err := WriteLog(w, log); err != nil {
			return err
		}
	}

	return nil
}

func ReadLogs(r io.Reader) ([]*Log, error) {
	var logs []*Log

	for {
		log, err := ReadLog(r)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, errors.Errorf("unable to unmarshal log entries: %v", err)
		}

		logs = append(logs, log)
	}

	return logs, nil
}

func WriteLog(w io.Writer, log *Log) error {
	// Convert log entry in efficient profobuf representation.
	entry, err := proto.Marshal(log)
	if err != nil {
		return errors.Errorf("unable to marshal log entry: %v", err)
	}

	// Encode the protobuf message length and write it.
	var l [2]byte
	length := uint16(len(entry))
	binary.BigEndian.PutUint16(l[:], length)

	if _, err := w.Write(l[:]); err != nil {
		return err
	}

	// Write the protobuf message itself.
	if _, err := w.Write(entry); err != nil {
		return err
	}

	return nil
}

func ReadLog(r io.Reader) (*Log, error) {
	var l [2]byte
	if _, err := r.Read(l[:]); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint16(l[:])
	entry := make([]byte, length)
	if _, err := r.Read(entry[:]); err != nil {
		return nil, err
	}

	log := &Log{}
	if err := proto.Unmarshal(entry, log); err != nil {
		return nil, errors.Errorf("unable to unmarshal log entry: %v", err)
	}

	return log, nil
}
