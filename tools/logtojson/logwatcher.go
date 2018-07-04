package main

import (
	hublogs "github.com/bitlum/hub/manager/logs"
	"github.com/go-errors/errors"
	"sync"
	"github.com/fsnotify/fsnotify"
	"os"
	"io"
)

type updateLogWatcher struct {
	*fsnotify.Watcher

	logs chan *hublogs.Log

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
		logs:    make(chan *hublogs.Log),
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

func (w *updateLogWatcher) subscribe() chan *hublogs.Log {
	return w.logs
}

func (w *updateLogWatcher) stop() {
	close(w.quit)
	w.Watcher.Close()
}

func (w *updateLogWatcher) sentLogs(r io.Reader) error {
	logs, err := hublogs.ReadLogs(r)
	if err != nil {
		return errors.Errorf("unable read logs: %v", err)
	}

	for _, log := range logs {
		w.logs <- log
	}

	return nil
}
