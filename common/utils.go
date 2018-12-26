package common

import (
	"encoding/json"
	"github.com/go-errors/errors"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// GetFunctionName() returns name of the function within which it executed.
func GetFunctionName() string {
	pc, _, _, _ := runtime.Caller(1)
	fullName := runtime.FuncForPC(pc).Name()
	parts := strings.Split(fullName, ".")
	return parts[len(parts)-1]
}

func NowInMilliSeconds() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func ConvertTimeToMilliSeconds(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}
func ConvertDurationToMilliSeconds(t time.Duration) int64 {
	return t.Nanoseconds() / int64(time.Millisecond)
}

func GetBitcoinUSDPRice() (float64, error) {
	resp, err := http.Get("https://blockchain.info/ticker")
	if err != nil {
		return 0, errors.Errorf("unable to fetch bitcoin price: %v",
			err)
	}
	defer resp.Body.Close()

	type Ticker struct {
		Last float64 `json:"last"`
	}

	type RespData map[string]Ticker
	d := json.NewDecoder(resp.Body)

	var data RespData
	if err := d.Decode(&data); err != nil {
		return 0, errors.Errorf("unable to decode bitcoin price: %v",
			err)
	}

	return data["USD"].Last, nil
}
