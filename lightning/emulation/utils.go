package emulation

import (
	"github.com/go-errors/errors"
)

func fail(errChan chan error, format string, params ...interface{}) error {
	err := errors.Errorf(format, params)
	log.Error(err)

	// Ensure that we are not sending in closed channel.
	select {
	case <-errChan:
		return err
	default:
	}

	select {
	case errChan <- err:
	default:
	}

	return err
}

func toMilli(satoshis int64) int64 {
	milSatInSat := int64(1000)
	return satoshis * milSatInSat
}

func fromMilli(milliSatoshis int64) int64 {
	milSatInSat := int64(1000)
	return milliSatoshis / milSatInSat
}
