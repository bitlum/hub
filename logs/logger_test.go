package logs

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

func TestEncodeDecode(t *testing.T) {
	entries1 := []*Log{
		{
			Time: time.Now().UnixNano(),
			Data: &Log_State{
				&NodeState{
					Channels: []*Channel{
						{
							UserId:        "1",
							RemoteBalance: 1,
							LocalBalance:  1,
						},
						{
							UserId:        "2",
							RemoteBalance: 2,
							LocalBalance:  2,
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
