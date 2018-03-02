package logger

import (
	"testing"
	"time"
	"bytes"
	"reflect"
)

func TestEncodeDecode(t *testing.T) {
	entries1 := []*Log{
		{
			Time: time.Now().UnixNano(),
			Data: &Log_State{
				&RouterState{
					Channels: []*Channel{
						{
							UserId:        1,
							UserBalance:   1,
							RouterBalance: 1,
						},
						{
							UserId:        2,
							UserBalance:   2,
							RouterBalance: 2,
						},
					},
					FreeBalance: 10,
				},
			},
		},
	}

	var b bytes.Buffer
	if err := WriteLogs(&b, entries1); err != nil {
		t.Fatal(err)
	}

	entries2, err := ReadLogs(&b)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(entries1, entries2) {
		t.Fatal("log entries are not the same")
	}
}
