package emulation

import "github.com/go-errors/errors"

func fail(errChan chan error, format string, params ...interface{}) error {
	err := errors.Errorf(format, params)
	log.Error(err)
	errChan <- err
	return err
}
