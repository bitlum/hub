package main

import (
	"github.com/bitlum/hub/manager/logger"
	"github.com/go-errors/errors"
	"sync"
	"github.com/fsnotify/fsnotify"
	"os"
	"io"
	"time"
)

type updateLogWatcher struct {
	*fsnotify.Watcher

	logs chan *logger.Log

	quit chan struct{}
	wg   sync.WaitGroup
}

func newUpdateLogWatcher(filePath string) (*updateLogWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Errorf("unable to init new watcher: %v", err)
	}

	err = watcher.Add(filePath)
	if err != nil {
		return nil, errors.Errorf("unable to add log file watcher: %v", err)
	}

	lw := &updateLogWatcher{
		quit:    make(chan struct{}),
		logs:    make(chan *logger.Log),
		Watcher: watcher,
	}

	// Open file and read initially existed logs
	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Errorf("unable open file: %v", err)
	}

	lw.wg.Add(1)
	go func() {
		defer func() {
			mainLog.Infof("Quiting update log watcher goroutine, file(%v)",
				filePath)
			lw.wg.Done()
		}()

		mainLog.Infof("Starting update log watcher goroutine, file(%v)",
			filePath)

		// Sent initially existed logs in the file
		if err := lw.sentLogs(f); err != nil {
			mainLog.Errorf("unable sent new logs: %v", err)
			return
		}

		for {
			select {
			case <-lw.quit:
				return

			case event := <-watcher.Events:
				if event.Op&fsnotify.Write != fsnotify.Write {
					continue
				}

				if err := lw.sentLogs(f); err != nil {
					mainLog.Errorf("unable sent new logs: %v", err)
					return
				}
			case err := <-watcher.Errors:
				mainLog.Errorf("watcher error: %v", err)
				close(lw.quit)
				return
			}
		}
	}()

	return lw, nil
}

func (w *updateLogWatcher) subscribe() chan *logger.Log {
	return w.logs
}

func (w *updateLogWatcher) stop() {
	close(w.quit)
	w.Watcher.Close()
}

func (w *updateLogWatcher) sentLogs(r io.Reader) error {
	logs, err := logger.ReadLogs(r)
	if err != nil {
		return errors.Errorf("unable read logs: %v", err)
	}

	for _, log := range logs {
		w.logs <- log
	}

	return nil
}

// wrapUpdateLogSubscription wraps updates subscription and slowed them
// down thereby reproducing updates in the same speed the were written in the
// log.
func wrapUpdateLogSubscription(updates chan *logger.Log) chan *logger.Log {
	var deltaNano int64
	var isDeltaInitialised bool

	reproducedUpdates := make(chan *logger.Log, 1000)

	go func() {
		for {
			update := <-updates
			now := time.Now().UnixNano()

			if !isDeltaInitialised {
				deltaNano = now - update.Time
				isDeltaInitialised = true
			}

			// If log was read faster that even has occurred,
			// than we should wait before reproducing update.
			waitNano := update.Time + deltaNano - now

			if waitNano > 0 {
				time.AfterFunc(time.Duration(waitNano), func() {
					reproducedUpdates <- update
				})
			} else {
				reproducedUpdates <- update
			}
		}
	}()

	return reproducedUpdates
}
